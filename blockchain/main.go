package blockchain

func CrateBlockchain() {
	bc := CreateBlockchain()
	defer bc.db.Close()
}
