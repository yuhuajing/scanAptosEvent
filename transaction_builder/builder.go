package transactionbuilder

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/coming-chat/lcs"
	"golang.org/x/crypto/sha3"
)

const (
	RAW_TRANSACTION_SALT           = "APTOS::RawTransaction"
	RAW_TRANSACTION_WITH_DATA_SALT = "APTOS::RawTransactionWithData"
)

type SigningMessage []byte

type Signable interface {
	GetSigningMessage() (SigningMessage, error)
}

func (t *RawTransaction) GetSigningMessage() (SigningMessage, error) {
	prefixBytes := sha3.Sum256([]byte(RAW_TRANSACTION_SALT))
	msg, err := lcs.Marshal(t)
	if err != nil {
		return nil, err
	}
	return append(prefixBytes[:], msg...), nil
}

func (t *MultiAgentRawTransaction) GetSigningMessage() (SigningMessage, error) {
	prefixBytes := sha3.Sum256([]byte(RAW_TRANSACTION_WITH_DATA_SALT))
	msg, err := lcs.Marshal(t)
	if err != nil {
		return nil, err
	}
	return append(prefixBytes[:], msg...), nil
}

// ------ TransactionBuilderEd25519 ------

type SigningFunctionEd25519 func(SigningMessage) []byte
type TransactionBuilderEd25519 struct {
	SigningFn SigningFunctionEd25519
	PublicKey []byte
}

func NewTransactionBuilderEd25519(signingFn SigningFunctionEd25519, publicKey []byte) *TransactionBuilderEd25519 {
	return &TransactionBuilderEd25519{signingFn, publicKey}
}

func (b *TransactionBuilderEd25519) Sign(rawTxn *RawTransaction) (data []byte, err error) {
	if b.SigningFn == nil {
		return nil, errors.New("Signing failed: you must specify a signing function")
	}
	signingMessage, err := rawTxn.GetSigningMessage()
	if err != nil {
		return
	}
	signatureBytes := b.SigningFn(signingMessage)
	if err != nil {
		return
	}
	publickey, err := NewEd25519PublicKey(b.PublicKey)
	if err != nil {
		return
	}
	signature, err := NewEd25519Signature(signatureBytes)
	if err != nil {
		return
	}
	authenticator := TransactionAuthenticatorEd25519{
		PublicKey: *publickey,
		Signature: *signature,
	}
	signedTxn := SignedTransaction{
		Transaction:   rawTxn,
		Authenticator: authenticator,
	}

	data, err = lcs.Marshal(signedTxn)
	return data, err
}

// ------ TransactionBuilderMultiEd25519 ------

type SigningFunctionMultiEd25519 func(SigningMessage) MultiEd25519Signature
type TransactionBuilderMultiEd25519 struct {
	SigningFn SigningFunctionMultiEd25519
	PublicKey MultiEd25519PublicKey
}

func (b *TransactionBuilderMultiEd25519) Sign(rawTxn *RawTransaction) (data []byte, err error) {
	if b.SigningFn == nil {
		return nil, errors.New("Signing failed: you must specify a signing function")
	}
	signingMessage, err := rawTxn.GetSigningMessage()
	if err != nil {
		return
	}
	signature := b.SigningFn(signingMessage)

	authenticator := TransactionAuthenticatorMultiEd25519{
		PublicKey: b.PublicKey,
		Signature: signature,
	}
	signedTxn := SignedTransaction{
		Transaction:   rawTxn,
		Authenticator: authenticator,
	}

	data, err = lcs.Marshal(signedTxn)
	return data, err
}

// ------ TransactionBuilderABI ------

type ABIBuilderConfig struct {
	Sender         AccountAddress
	SequenceNumber uint64
	GasUnitPrice   uint64
	MaxGasAmount   uint64
	ExpSecFromNow  uint64
	ChainId        uint8
}

type TransactionBuilderABI struct {
	ABIMap map[string]ScriptABI
	// BuildereConfig ABIBuilderConfig
}

func NewTransactionBuilderABI(abis [][]byte) (*TransactionBuilderABI, error) {
	abiMap := make(map[string]ScriptABI)
	for _, bytes := range abis {
		var abi ScriptABI
		err := lcs.Unmarshal(bytes, &abi)
		if err != nil {
			return nil, err
		}

		k := ""
		if funcABI, ok := abi.(EntryFunctionABI); ok {
			module := funcABI.ModuleName
			k = fmt.Sprintf("%v::%v::%v", module.Address.ToShortString(), module.Name, funcABI.Name)
		} else {
			funcABI := abi.(TransactionScriptABI)
			k = funcABI.Name
		}

		if abiMap[k] != nil {
			return nil, errors.New("Found conflicting ABI interfaces")
		}
		abiMap[k] = abi
	}

	return &TransactionBuilderABI{
		ABIMap: abiMap,
	}, nil
}

func (tb *TransactionBuilderABI) BuildTransactionPayload(function string, tyTags []string, args []any) (TransactionPayload, error) {
	tag, err := NewTypeTagStructFromString(function)
	if err != nil {
		return nil, fmt.Errorf("Invalid function: %v", function)
	}
	function = fmt.Sprintf("%v::%v::%v", tag.Address.ToShortString(), tag.ModuleName, tag.Name)
	scriptABI, ok := tb.ABIMap[function]
	if !ok {
		return nil, fmt.Errorf("Cannot find function: %v", function)
	}

	typeTags := []TypeTag{}
	for _, tagString := range tyTags {
		parser, err := NewTypeTagParser(tagString)
		if err != nil {
			return nil, err
		}
		tag, err := parser.ParseTypeTag()
		if err != nil {
			return nil, err
		}
		typeTags = append(typeTags, tag)
	}

	var payload TransactionPayload
	if funcABI, ok := scriptABI.(EntryFunctionABI); ok {
		bcsArgs, err := toBCSArgs(funcABI.Args, args)
		if err != nil {
			return nil, err
		}
		payload = TransactionPayloadEntryFunction{
			ModuleName:   funcABI.ModuleName,
			FunctionName: Identifier(funcABI.Name),
			TyArgs:       typeTags,
			Args:         bcsArgs,
		}
	} else if funcABI, ok := scriptABI.(TransactionScriptABI); ok {
		scriptArgs, err := toTransactionArguments(funcABI.Args, args)
		if err != nil {
			return nil, err
		}
		payload = TransactionPayloadScript{
			Code:   funcABI.Code,
			TyArgs: typeTags,
			Args:   scriptArgs,
		}
	} else {
		return nil, errors.New("Unsupported script abi.")
	}

	return TransactionPayload(payload), nil
}

func toBCSArgs(abiArgs []ArgumentABI, args []any) ([][]byte, error) {
	if len(abiArgs) != len(args) {
		return nil, errors.New("Wrong number of args provided.")
	}

	res := [][]byte{}
	for i, arg := range args {
		var b bytes.Buffer
		encoder := lcs.NewEncoder(&b)
		err := serializeArg(arg, abiArgs[i].TypeTag, encoder)
		if err != nil {
			return nil, err
		}
		res = append(res, b.Bytes())
	}

	return res, nil
}

func toTransactionArguments(abiArgs []ArgumentABI, args []any) ([]TransactionArgument, error) {
	if len(abiArgs) != len(args) {
		return nil, errors.New("Wrong number of args provided.")
	}

	res := []TransactionArgument{}
	for i, arg := range args {
		argument, err := argToTransactionArgument(arg, abiArgs[i].TypeTag)
		if err != nil {
			return nil, err
		}
		res = append(res, argument)
	}
	return res, nil
}
