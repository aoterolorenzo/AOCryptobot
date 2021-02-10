package common

import (
	"../../model"
	"../../service/binance"
	"fmt"
)

type WalletService struct {
	Wallet *model.Wallet
	Coin1  string
	Coin2  string
}

func (ws *WalletService) GetTotalAssetsBalance(price float64) float64 {
	return (ws.Wallet.Coin1FreeBalance+ws.Wallet.Coin1LockedBalance)*price +
		(ws.Wallet.Coin2FreeBalance + ws.Wallet.Coin2LockedBalance)
}

func (ws *WalletService) GetAssetBalance(coin string) (float64, error) {
	if coin == ws.Coin1 {
		return ws.Wallet.Coin1FreeBalance + ws.Wallet.Coin1LockedBalance, nil
	}

	if coin == ws.Coin2 {
		return ws.Wallet.Coin2FreeBalance + ws.Wallet.Coin2LockedBalance, nil
	}

	return -1.0, fmt.Errorf("error: unavailable pair provided")
}

func (ws *WalletService) InitWallet() {
	ws.Wallet = &model.Wallet{}
}

func (ws *WalletService) UpdateWallet() error {
	marketService := binance.BinanceService{}
	marketService.SetPair(ws.Coin1 + ws.Coin2)
	marketService.ConfigureClient()

	coin1FreeBalance, err := marketService.GetAvailableBalance(ws.Coin1)
	if err != nil {
		return err
	}
	ws.Wallet.Coin1FreeBalance = coin1FreeBalance

	coin2FreeBalance, err := marketService.GetAvailableBalance(ws.Coin2)
	if err != nil {
		return err
	}
	ws.Wallet.Coin2FreeBalance = coin2FreeBalance

	coin1LockedBalance, err := marketService.GetLockedBalance(ws.Coin1)
	if err != nil {
		return err
	}
	ws.Wallet.Coin1LockedBalance = coin1LockedBalance

	coin2LockedBalance, err := marketService.GetLockedBalance(ws.Coin2)
	if err != nil {
		return err
	}
	ws.Wallet.Coin2LockedBalance = coin2LockedBalance

	return nil
}