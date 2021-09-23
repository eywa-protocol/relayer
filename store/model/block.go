package model



type Block struct {
     bm BaseModel
	 tx []Transaction
	 chanId int
}

type BaseModel struct {
	id int
}



