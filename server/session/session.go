package session

/*
Copyright (c) 2019 m15o <m15o@posteo.net> . All rights reserved.
Copyright (c) 2022 cblgh <source-code@cblgh.org> . All rights reserved.

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE
USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

import (
	"gomod.cblgh.org/cerca/util/eout"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

type Session struct {
	cookieName string
	Store           *sessions.CookieStore
	ShortLivedStore *sessions.CookieStore
}

func New(authKey, cookieName string, developing bool) *Session {
	store := sessions.NewCookieStore([]byte(authKey))
	store.Options = &sessions.Options{
		HttpOnly: true,
		Secure:   !developing,
		MaxAge:   86400 * 30,
	}
	short := sessions.NewCookieStore([]byte(authKey))
	short.Options = &sessions.Options{
		HttpOnly: true,
		Secure:   !developing,
		MaxAge: 600, // 10 minutes
	}
	return &Session{
		Store:           store,
		ShortLivedStore: short,
		cookieName: cookieName,
	}
}

func (s *Session) Delete(res http.ResponseWriter, req *http.Request) error {
	ed := eout.Describe("delete session cookie")
	clearSession := func(store *sessions.CookieStore) error {
		session, err := store.Get(req, s.cookieName)
		if err != nil {
			return ed.Eout(err, "get session")
		}
		session.Options.MaxAge = -1
		err = session.Save(req, res)
		return ed.Eout(err, "save expired session")
	}
	err := clearSession(s.Store)
	if err != nil {
		return err
	}
	err = clearSession(s.ShortLivedStore)
	return err
}

func (s *Session) getValueFromSession(req *http.Request, store *sessions.CookieStore, key string) (interface{}, error) {
	session, err := store.Get(req, s.cookieName)
	if err != nil {
		return nil, err
	}
	value, ok := session.Values[key]
	if !ok {
		err := errors.New(fmt.Sprintf("extracting %s from session; no such value", key))
		return nil, eout.Eout(err, "get session")
	}
	return value, nil
}


func (s *Session) GetInt(req *http.Request, key string) (int, error) {
	val, err := s.getValueFromSession(req, s.Store, key)
	if val == nil || err != nil {
		return -1, err
	}
	return val.(int), err
}

/* TODO (2024-11-20): revamp structure of this file to something less repetitive and using enum-like things instead */
func (s *Session) GetString(req *http.Request, key string) (string, error) {
	val, err := s.getValueFromSession(req, s.Store, key)
	if val == nil || err != nil {
		return "", err
	}
	return val.(string), err
}

func (s *Session) genericSave(req *http.Request, res http.ResponseWriter, shortLived bool, key string, val interface{}) error {
	store := s.Store
	if shortLived {
		store = s.ShortLivedStore
	}
	session, _ := store.Get(req, s.cookieName)
	session.Values[key] = val
	return session.Save(req, res)
}

func (s *Session) SaveInt(req *http.Request, res http.ResponseWriter, key string, val int) error {
	return s.genericSave(req, res, false, key, val)
}

func (s *Session) SaveString(req *http.Request, res http.ResponseWriter, key, val string) error {
	return s.genericSave(req, res, false, key, val)
}
