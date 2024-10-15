package store

import (
	"net/url"
	"strings"

	"github.com/go-oauth2/oauth2/v4/errors"
)

// ValidateURIHandler validates the redirectURI used by the manager.
// mg.SetValidateURIHandler(dbStore.ValidateURIHandler).
func (*Store) ValidateURIHandler(baseURI string, redirectURIs string) (err error) {
	base, err := url.Parse(baseURI)
	if err != nil {
		return err
	}

	// Since the oauth2 package only supports one redirectURI, we need to split the string and check each one
	uriList := strings.Split(redirectURIs, ",")
	for _, redirectURI := range uriList {
		redirect, err := url.Parse(redirectURI)
		if err != nil {
			return err
		}

		// Check if the redirectURI is a subset of the baseURI
		if strings.HasSuffix(redirect.Host, base.Host) {
			return nil
		}
	}
	return errors.ErrInvalidRedirectURI
}
