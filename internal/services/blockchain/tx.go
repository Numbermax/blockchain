package blockchain

type TxOutput struct {
	Value     int
	PublicKey string
}

func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PublicKey == data
}

type TxInput struct {
	ID  []byte
	Out int
	Sig string
}

func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}
