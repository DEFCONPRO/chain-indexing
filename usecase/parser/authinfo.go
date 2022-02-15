package parser

import (
	"fmt"
	"strconv"

	"github.com/crypto-com/chain-indexing/external/tmcosmosutils"
	"github.com/crypto-com/chain-indexing/internal/base64"
	"github.com/crypto-com/chain-indexing/usecase/model"
	"github.com/crypto-com/chain-indexing/usecase/parser/utils"
)

func ParseSignerInfosToTransactionSigners(
	signerInfos []utils.SignerInfo,
	accountAddressPrefix string,
) ([]model.TransactionSigner, error) {
	var signers []model.TransactionSigner

	for _, signer := range signerInfos {
		var transactionSignerInfo *model.TransactionSignerKeyInfo
		var address string

		sequence, parseErr := strconv.ParseUint(signer.Sequence, 10, 64)
		if parseErr != nil {
			return nil, fmt.Errorf("error parsing account sequence: %v", parseErr)
		}
		if signer.ModeInfo.MaybeSingle != nil {
			if signer.MaybePublicKey == nil {
				// FIXME: extract signer address from message: https://github.com/crypto-com/chain-indexing/issues/685
				address = ""
			} else {
				transactionSignerInfo = &model.TransactionSignerKeyInfo{
					Type:       signer.MaybePublicKey.Type,
					IsMultiSig: false,
					Pubkeys:    []string{*signer.MaybePublicKey.MaybeKey},
				}

				parsedAddr, parseAddrErr := ParseTransactionSignerInfoToAddress(*transactionSignerInfo, accountAddressPrefix)
				if parseAddrErr != nil {
					return nil, fmt.Errorf("error parsing signer info to address: %v", parseAddrErr)
				}
				address = parsedAddr
			}
		} else {
			pubkeys := make([]string, 0, len(signer.MaybePublicKey.MaybePublicKeys))
			for _, pubkey := range signer.MaybePublicKey.MaybePublicKeys {
				pubkeys = append(pubkeys, pubkey.Key)
			}
			transactionSignerInfo = &model.TransactionSignerKeyInfo{
				Type:           signer.MaybePublicKey.Type,
				IsMultiSig:     true,
				Pubkeys:        pubkeys,
				MaybeThreshold: signer.MaybePublicKey.MaybeThreshold,
			}

			parsedAddr, parseAddrErr := ParseTransactionSignerInfoToAddress(*transactionSignerInfo, accountAddressPrefix)
			if parseAddrErr != nil {
				return nil, fmt.Errorf("error parsing signer info to address: %v", parseAddrErr)
			}
			address = parsedAddr
		}

		signers = append(signers, model.TransactionSigner{
			MaybeKeyInfo:    transactionSignerInfo,
			Address:         address,
			AccountSequence: sequence,
		})
	}

	return signers, nil
}

func ParseTransactionSignerInfoToAddress(
	signerInfo model.TransactionSignerKeyInfo,
	accountAddressPrefix string,
) (string, error) {
	var address string
	if signerInfo.IsMultiSig {
		addrPubKeys := make([][]byte, 0, len(signerInfo.Pubkeys))
		for _, pubKey := range signerInfo.Pubkeys {
			rawPubKey := base64.MustDecodeString(pubKey)
			addrPubKeys = append(addrPubKeys, rawPubKey)
		}
		var multiSigAddrErr error
		address, multiSigAddrErr = tmcosmosutils.MultiSigAddressFromPubKeys(
			accountAddressPrefix,
			addrPubKeys,
			*signerInfo.MaybeThreshold,
			false,
		)
		if multiSigAddrErr != nil {
			return "", fmt.Errorf("error converting public keys to multisig address: %v", multiSigAddrErr)
		}
	} else {
		var addrErr error
		pubKey := base64.MustDecodeString(signerInfo.Pubkeys[0])
		address, addrErr = tmcosmosutils.AccountAddressFromPubKey(accountAddressPrefix, pubKey)
		if addrErr != nil {
			return "", fmt.Errorf("error converting public key to address: %v", addrErr)
		}
	}
	return address, nil
}
