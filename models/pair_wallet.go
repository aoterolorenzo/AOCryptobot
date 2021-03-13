package models

type PairWallet struct {
	Coin1FreeBalance   float64
	Coin1LockedBalance float64
	Coin2FreeBalance   float64
	Coin2LockedBalance float64
}

func NewPairWallet() PairWallet {
	return PairWallet{}
}
