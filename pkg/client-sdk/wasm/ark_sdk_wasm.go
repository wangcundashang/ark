package arksdkwasm

import (
	"context"
	"fmt"
	"strings"
	"syscall/js"

	arksdk "github.com/ark-network/ark-sdk"
	"github.com/ark-network/ark-sdk/store"
	"github.com/ark-network/ark-sdk/wallet"
	singlekeywallet "github.com/ark-network/ark-sdk/wallet/singlekey"
	walletstore "github.com/ark-network/ark-sdk/wallet/singlekey/store"
)

var (
	arkSdkClient arksdk.ArkClient
	configStore  store.ConfigStore
)

func init() {
	js.Global().Set("init", InitWrapper())
	js.Global().Set("unlock", UnlockWrapper())
	js.Global().Set("lock", LockWrapper())
	js.Global().Set("balance", BalanceWrapper())
	js.Global().Set("onboard", OnboardWrapper())
	js.Global().Set("receive", ReceiveWrapper())
	js.Global().Set("sendOnChain", SendOnChainWrapper())
	js.Global().Set("sendOffChain", SendOffChainWrapper())
	js.Global().Set("unilateralRedeem", UnilateralRedeemWrapper())
	js.Global().Set("collaborativeRedeem", CollaborativeRedeemWrapper())
	js.Global().Set("log", LogWrapper())

	js.Global().Set("getAspUrl", GetAspUrlWrapper())
	js.Global().Set("getAspPubKeyHex", GetAspPubkeyWrapper())
	js.Global().Set("getWalletType", GetWalletTypeWrapper())
	js.Global().Set("getClientType", GetClientTypeWrapper())
	js.Global().Set("getNetwork", GetNetworkWrapper())
	js.Global().Set("getRoundLifetime", GetRoundLifetimeWrapper())
	js.Global().Set("getUnilateralExitDelay", GetUnilateralExitDelayWrapper())
	js.Global().Set("getMinRelayFee", GetMinRelayFeeWrapper())
}

func NewCovenantClient(
	ctx context.Context, storeSvc store.ConfigStore,
) error {
	var err error

	data, err := storeSvc.GetData(ctx)
	if err != nil {
		return err
	}

	if data == nil {
		arkSdkClient, err = arksdk.NewCovenantClient(storeSvc)
	} else {
		var walletSvc wallet.WalletService
		switch data.WalletType {
		case arksdk.SingleKeyWallet:
			walletSvc, err = getSingleKeyWallet(storeSvc, data.Network.Name)
			if err != nil {
				return err
			}
		// TODO: Support HD wallet
		default:
			return fmt.Errorf("unknown wallet type")
		}
		arkSdkClient, err = arksdk.LoadCovenantClientWithWallet(storeSvc, walletSvc)
	}
	if err != nil {
		js.Global().Get("console").Call("error", err.Error())
		return err
	}
	configStore = storeSvc

	select {}
}

func NewCovenantlessClient(
	ctx context.Context, storeSvc store.ConfigStore,
) error {
	var err error

	data, err := storeSvc.GetData(ctx)
	if err != nil {
		return err
	}

	if data == nil {
		arkSdkClient, err = arksdk.NewCovenantlessClient(storeSvc)
	} else {
		var walletSvc wallet.WalletService
		switch data.WalletType {
		case arksdk.SingleKeyWallet:
			walletSvc, err = getSingleKeyWallet(storeSvc, data.Network.Name)
			if err != nil {
				return err
			}
		// TODO: Support HD wallet
		default:
			return fmt.Errorf("unknown wallet type")
		}
		arkSdkClient, err = arksdk.LoadCovenantlessClientWithWallet(storeSvc, walletSvc)
	}
	if err != nil {
		js.Global().Get("console").Call("error", err.Error())
		return err
	}
	configStore = storeSvc

	select {}
}

func getWalletStore(storeType string) (walletstore.WalletStore, error) {
	if storeType == LocalStorageStore {
		return NewLocalStorageWalletStore()
	}
	// TODO: Support IndexDB store
	return nil, fmt.Errorf("unknown wallet store type")
}

func getSingleKeyWallet(
	configStore store.ConfigStore, network string,
) (wallet.WalletService, error) {
	walletStore, err := getWalletStore(configStore.GetType())
	if err != nil {
		return nil, err
	}
	if strings.Contains(network, "liquid") {
		return singlekeywallet.NewLiquidWallet(configStore, walletStore)
	}
	return singlekeywallet.NewBitcoinWallet(configStore, walletStore)
}
