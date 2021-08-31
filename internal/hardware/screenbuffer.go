package hardware

type Screenbuffer struct {
	pixels []int
	xDim   int
	yDim   int
}

func NewScreenbuffer(x, y int) *Screenbuffer {
	return &Screenbuffer{make([]int, x*y), x, y}
}

func (sb *Screenbuffer) Put(x, y, value int) bool {
	willUnset := sb.pixels[x*sb.yDim+y] != 0 && value == 0
	sb.pixels[x*sb.yDim+y] = value
	return willUnset
}

func (sb *Screenbuffer) Get(x, y int) int {
	return sb.pixels[x*sb.yDim+y]
}
