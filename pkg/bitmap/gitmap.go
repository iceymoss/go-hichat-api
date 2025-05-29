package bitmap

type Bitmap struct {
	bits []byte
	size int
}

func NewBitmap(size int) *Bitmap {
	if size == 0 {
		size = 256
	}
	return &Bitmap{
		bits: make([]byte, size),
		size: size * 8,
	}
}

func (b *Bitmap) Set(id string) {
	// id在那个bit
	idx := hash(id) % b.size
	// 计算在那个byte
	byteIdx := idx / 8
	// 在这个byte中的那个bit位置
	bitIdx := idx % 8

	b.bits[byteIdx] |= 1 << bitIdx
}

func (b *Bitmap) IsSet(id string) bool {
	idx := hash(id) % b.size
	// 计算在那个byte
	byteIdx := idx / 8
	// 在这个byte中的那个bit位置
	bitIdx := idx % 8
	return (b.bits[byteIdx] & (1 << bitIdx)) != 0
}

func (b *Bitmap) Export() []byte {
	return b.bits
}
func Load(bits []byte) *Bitmap {
	if len(bits) == 0 {
		return NewBitmap(0)
	}

	return &Bitmap{
		bits: bits,
		size: len(bits) * 8,
	}
}

func hash(id string) int {
	// 使用BKDR哈希算法
	seed := 131313 // 31 131 1313 13131 131313, etc
	h := 0
	for _, c := range id {
		h = h*seed + int(c)
	}
	return h & 0x7FFFFFFF
}
