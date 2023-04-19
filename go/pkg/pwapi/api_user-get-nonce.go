package pwapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"pathwar.land/pathwar/v2/go/pkg/errcode"
	"pathwar.land/pathwar/v2/go/pkg/pwdb"
)

func (svc *service) UserGetNonce(ctx context.Context, in *UserGetNonce_Input) (*UserGetNonce_Output, error) {
	if in == nil || in.Address == "" {
		return nil, errcode.ErrMissingInput
	}

	nonce, err := GenerateNonce()
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	var user pwdb.User
	err = svc.db.
		Where(pwdb.User{
			Username: in.Address,
		}).Find(&user).Error
	if err != nil {
		fmt.Println("THIS IS err", err)
		user = pwdb.User{
			Username: in.Address,
			Slug:     in.Address,
			Nonce:    nonce,
		}
		err = svc.db.Create(&user).Error
	} else {
		user.Nonce = nonce
		err = svc.db.Update(&user).Error
	}

	ret := &UserGetNonce_Output{Nonce: nonce}
	return ret, nil
}

func GenerateNonce() (string, error) {
	nonceBytes := make([]byte, 32)
	_, err := rand.Read(nonceBytes)
	if err != nil {
		return "", fmt.Errorf("could not generate nonce")
	}

	return base64.URLEncoding.EncodeToString(nonceBytes), nil
}
