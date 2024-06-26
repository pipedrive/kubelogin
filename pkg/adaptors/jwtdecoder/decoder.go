// Package jwtdecoder provides decoding a JWT.
package jwtdecoder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/wire"
	"golang.org/x/xerrors"
)

//go:generate mockgen -destination mock_jwtdecoder/mock_jwtdecoder.go github.com/int128/kubelogin/pkg/adaptors/jwtdecoder Interface

// Set provides an implementation and interface.
var Set = wire.NewSet(
	wire.Struct(new(Decoder), "*"),
	wire.Bind(new(Interface), new(*Decoder)),
)

type Interface interface {
	Decode(s string) (*Claims, error)
}

// Claims represents claims of a token.
type Claims struct {
	Subject string
	Expiry  time.Time
	Pretty  map[string]string // string representation for debug and logging
}

type Decoder struct{}

// Decode returns the claims of the JWT.
// Note that this method does not verify the signature and always trust it.
func (d *Decoder) Decode(s string) (*Claims, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return nil, xerrors.Errorf("token contains an invalid number of segments")
	}
	//nolint
	b, err := jwt.DecodeSegment(parts[1])
	if err != nil {
		return nil, xerrors.Errorf("could not decode the token: %w", err)
	}
	//nolint
	var claims jwt.StandardClaims
	if err := json.NewDecoder(bytes.NewBuffer(b)).Decode(&claims); err != nil {
		return nil, xerrors.Errorf("could not decode the json of token: %w", err)
	}
	var rawClaims map[string]interface{}
	if err := json.NewDecoder(bytes.NewBuffer(b)).Decode(&rawClaims); err != nil {
		return nil, xerrors.Errorf("could not decode the json of token: %w", err)
	}
	return &Claims{
		Subject: claims.Subject,
		Expiry:  time.Unix(claims.ExpiresAt, 0),
		Pretty:  dumpRawClaims(rawClaims),
	}, nil
}

func dumpRawClaims(rawClaims map[string]interface{}) map[string]string {
	claims := make(map[string]string)
	for k, v := range rawClaims {
		switch v := v.(type) {
		case float64:
			claims[k] = fmt.Sprintf("%.f", v)
		default:
			claims[k] = fmt.Sprintf("%v", v)
		}
	}
	return claims
}
