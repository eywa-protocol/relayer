package model


type Block struct {
	header Header
	 bm      BaseModel
	 tx        []Transaction
	 chanId int
}

type Header struct {}

func CopyHeader(header Header)  *Header{
	return &header
}

// Header returns a copy of Header.
func (b *Block) Header() *Header { return CopyHeader(b.header) }

type BaseModel struct {
	id int
}



