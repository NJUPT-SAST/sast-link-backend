package store

import (
	"net/url"
	"strings"

	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/go-oauth2/oauth2/v4/errors"
)

func (s *Store) ValidateURIHandler(baseURI string, redirectURIs string) (err error) {
	base, err := url.Parse(baseURI)
	if err != nil {
		return err
	}
	log.Debugf("BaseURI: %s, RedirectURIs: %s", baseURI, redirectURIs)

	// Since the oauth2 package only supports one redirectURI, we need to split the string and check each one
	uriList := strings.Split(redirectURIs, ",")
	for _, redirectURI := range uriList {
		redirect, err := url.Parse(redirectURI)
		if err != nil {
			return err
		}

		log.Debugf("Base Host: %s, Redirect Host: %s", base.Host, redirect.Host)
		// Check if the redirectURI is a subset of the baseURI
		if strings.HasSuffix(redirect.Host, base.Host) {
			return nil
		}
	}
	return errors.ErrInvalidRedirectURI
}
