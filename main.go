package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

func parseReader(bin string, r io.Reader) ([]byte, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeCharDevice != 0 {
		return nil, fmt.Errorf("Invalid pipe data.\n\nUsage: echo \"Error(string)\" | %s", bin)
	}
	reader := bufio.NewReader(r)
	return ioutil.ReadAll(reader)
}

func sender(tx *types.Transaction) (common.Address, error) {
	signer := types.LatestSignerForChainID(tx.ChainId())
	return types.Sender(signer, tx)
}

func prettyTx(tx *types.Transaction) (string, error) {
	sender, err := sender(tx)
	if err != nil {
		return "", err
	}
	v, r, s := tx.RawSignatureValues()
	pretty := fmt.Sprintf(`
  Hash:      %s
  ChainID:   %s
  Type:      %d
  From:      %s
  To:        %s
  Nonce:     %d
  GasPrice:  %s
  GasLimit:  %d
    v:       %x
    r:       %x
    s:       %x
  Value:     %s
`,
		tx.Hash().Hex(),
		tx.ChainId(),
		tx.Type(),
		sender.Hex(),
		tx.To().Hex(),
		tx.Nonce(),
		tx.GasPrice(),
		tx.Gas(),
		v,
		r,
		s,
		tx.Value(),
	)
	d := tx.Data()
	if len(d) > 0 {
		return pretty + fmt.Sprintf("   Data:     %x", d), nil
	}
	return pretty, nil
}

func main() {
	info, err := os.Stdin.Stat()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(os.Args) > 0 && info.Size() > 0 {
		fmt.Fprintln(os.Stderr, "Accepts either piped data or a single arg, but not both")
		os.Exit(1)
	}
	var b []byte
	if len(os.Args) > 1 {
		b = []byte(os.Args[1])
	} else {
		b, err = parseReader(os.Args[0], os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		b = bytes.TrimSpace(b)
	}
	raw, err := hex.DecodeString(string(b))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var tx types.Transaction
	err = rlp.DecodeBytes(raw, &tx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	s, err := prettyTx(&tx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stdout, s)
}
