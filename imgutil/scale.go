// this file provides a modified subset of x/image/draw optimized for resampling of grayscale
// images.

package imgutil

import (
	"image"
	"math"
	"sync"
)

type cacheScaler struct {
	kernel *Kernel
	cache  map[cacheKey]Scaler
	mu     sync.Mutex
}

type cacheKey struct {
	dw, dh, sw, sh int
}

// Scale implements the Scaler interface.
func (z *cacheScaler) Scale(dst, src *image.Gray) {
	key := cacheKey{dst.Rect.Dx(), dst.Rect.Dy(), src.Rect.Dx(), src.Rect.Dy()}

	z.mu.Lock()
	scaler, ok := z.cache[key]
	if !ok {
		scaler = z.kernel.NewScaler(key.dw, key.dh, key.sw, key.sh)
		z.cache[key] = scaler
	}
	z.mu.Unlock()

	scaler.Scale(dst, src)
}

// NewCacheScaler creates, caches and reuses kernel scalers optimized for each unique combination
// of destination and source width and height. It is mostly useful when scaling a large quantity of
// images in a few fixed sizes.
func NewCacheScaler(k *Kernel) Scaler {
	return &cacheScaler{
		kernel: k,
		cache:  make(map[cacheKey]Scaler),
		mu:     sync.Mutex{},
	}
}

// Scaler scales the source image to the destination image.
type Scaler interface {
	Scale(dst, src *image.Gray)
}

// Kernel is an interpolator that blends source pixels weighted by a symmetric
// kernel function.
type Kernel struct {
	// Support is the kernel support and must be >= 0. At(t) is assumed to be
	// zero when t >= Support.
	Support float64
	// At is the kernel function. It will only be called with t in the
	// range [0, Support).
	At func(t float64) float64
}

// Scale implements the Scaler interface.
func (q *Kernel) Scale(dst, src *image.Gray) {
	q.newScaler(dst.Rect.Dx(), dst.Rect.Dy(), src.Rect.Dx(), src.Rect.Dy(), false).Scale(dst, src)
}

// NewScaler returns a Scaler that is optimized for scaling multiple times with
// the same fixed destination and source width and height.
func (q *Kernel) NewScaler(dw, dh, sw, sh int) Scaler {
	return q.newScaler(dw, dh, sw, sh, true)
}

func (q *Kernel) newScaler(dw, dh, sw, sh int, usePool bool) Scaler {
	s := &kernelScaler{
		kernel:     q,
		dw:         int32(dw),
		dh:         int32(dh),
		sw:         int32(sw),
		sh:         int32(sh),
		horizontal: newDistrib(q, int32(dw), int32(sw)),
		vertical:   newDistrib(q, int32(dh), int32(sh)),
	}
	if usePool {
		s.pool.New = func() interface{} {
			tmp := s.makeTmpBuf()
			return &tmp
		}
	}
	return s
}

var (
	// CatmullRom is the Catmull-Rom kernel. It is very slow, but usually gives
	// very high quality results.
	//
	// It is an instance of the more general cubic BC-spline kernel with parameters
	// B=0 and C=0.5. See Mitchell and Netravali, "Reconstruction Filters in
	// Computer Graphics", Computer Graphics, Vol. 22, No. 4, pp. 221-228.
	CatmullRom = &Kernel{2, func(t float64) float64 {
		if t < 1 {
			return (1.5*t-2.5)*t*t + 1
		}
		return ((-0.5*t+2.5)*t-4)*t + 2
	}}
)

type kernelScaler struct {
	kernel               *Kernel
	dw, dh, sw, sh       int32
	horizontal, vertical distrib
	pool                 sync.Pool
}

func (z *kernelScaler) makeTmpBuf() []float64 {
	return make([]float64, z.dw*z.sh)
}

// source is a range of contribs, their inverse total weight, and that ITW
// divided by 0xffff.
type source struct {
	i, j               int32
	invTotalWeight     float64
	invTotalWeightFFFF float64
}

// contrib is the weight of a column or row.
type contrib struct {
	coord  int32
	weight float64
}

// distrib measures how source pixels are distributed over destination pixels.
type distrib struct {
	// sources are what contribs each column or row in the source image owns,
	// and the total weight of those contribs.
	sources []source
	// contribs are the contributions indexed by sources[s].i and sources[s].j.
	contribs []contrib
}

// newDistrib returns a distrib that distributes sw source columns (or rows)
// over dw destination columns (or rows).
func newDistrib(q *Kernel, dw, sw int32) distrib {
	scale := float64(sw) / float64(dw)
	halfWidth, kernelArgScale := q.Support, 1.0
	// When shrinking, broaden the effective kernel support so that we still
	// visit every source pixel.
	if scale > 1 {
		halfWidth *= scale
		kernelArgScale = 1 / scale
	}

	// Make the sources slice, one source for each column or row, and temporarily
	// appropriate its elements' fields so that invTotalWeight is the scaled
	// coordinate of the source column or row, and i and j are the lower and
	// upper bounds of the range of destination columns or rows affected by the
	// source column or row.
	n, sources := int32(0), make([]source, dw)
	for x := range sources {
		center := (float64(x)+0.5)*scale - 0.5
		i := int32(math.Floor(center - halfWidth))
		if i < 0 {
			i = 0
		}
		j := int32(math.Ceil(center + halfWidth))
		if j > sw {
			j = sw
			if j < i {
				j = i
			}
		}
		sources[x] = source{i: i, j: j, invTotalWeight: center}
		n += j - i
	}

	contribs := make([]contrib, 0, n)
	for k, b := range sources {
		totalWeight := 0.0
		l := int32(len(contribs))
		for coord := b.i; coord < b.j; coord++ {
			t := abs((b.invTotalWeight - float64(coord)) * kernelArgScale)
			if t >= q.Support {
				continue
			}
			weight := q.At(t)
			if weight == 0 {
				continue
			}
			totalWeight += weight
			contribs = append(contribs, contrib{coord, weight})
		}
		totalWeight = 1 / totalWeight
		sources[k] = source{
			i:                  l,
			j:                  int32(len(contribs)),
			invTotalWeight:     totalWeight,
			invTotalWeightFFFF: totalWeight / 0xffff,
		}
	}

	return distrib{sources, contribs}
}

// abs is like math.Abs, but it doesn't care about negative zero, infinities or
// NaNs.
func abs(f float64) float64 {
	if f < 0 {
		f = -f
	}
	return f
}

// ftou converts the range [0.0, 1.0] to [0, 0xffff].
func ftou(f float64) uint16 {
	i := int32(0xffff*f + 0.5)
	if i > 0xffff {
		return 0xffff
	}
	if i > 0 {
		return uint16(i)
	}
	return 0
}

func (z *kernelScaler) Scale(dst, src *image.Gray) {
	if z.dw != int32(dst.Rect.Dx()) ||
		z.dh != int32(dst.Rect.Dy()) ||
		z.sw != int32(src.Rect.Dx()) ||
		z.sh != int32(src.Rect.Dy()) {
		z.kernel.Scale(dst, src)
		return
	}

	// Create a temporary buffer:
	// scaleX distributes the source image's columns over the temporary image.
	// scaleY distributes the temporary image's rows over the destination image.
	var tmp []float64
	if z.pool.New != nil {
		tmpp := z.pool.Get().(*[]float64)
		defer z.pool.Put(tmpp)
		tmp = *tmpp
	} else {
		tmp = z.makeTmpBuf()
	}

	z.scaleX(tmp, src)
	z.scaleY(dst, tmp)
}

func (z *kernelScaler) scaleX(tmp []float64, src *image.Gray) {
	t := 0
	for y := int32(0); y < z.sh; y++ {
		for _, s := range z.horizontal.sources {
			var p float64
			for _, c := range z.horizontal.contribs[s.i:s.j] {
				pi := int(y)*src.Stride + int(c.coord)
				pru := uint32(src.Pix[pi]) * 0x101
				p += float64(pru) * c.weight
			}
			p *= s.invTotalWeightFFFF
			tmp[t] = p
			t++
		}
	}
}

func (z *kernelScaler) scaleY(dst *image.Gray, tmp []float64) {
	for dx := int32(dst.Rect.Min.X); dx < int32(dst.Rect.Max.X); dx++ {
		d := int(dx)
		for _, s := range z.vertical.sources[dst.Rect.Min.Y:dst.Rect.Max.Y] {
			var p float64
			for _, c := range z.vertical.contribs[s.i:s.j] {
				p += tmp[c.coord*z.dw+dx] * c.weight
			}
			dst.Pix[d] = uint8(ftou(p*s.invTotalWeight) >> 8)
			d += dst.Stride
		}
	}
}
