package session

/*
Copyright (c) 2019 m15o <m15o@posteo.net> . All rights reserved.
Copyright (c) 2022 cblgh <m15o@posteo.net> . All rights reserved.

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
	"cerca/util"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

const cookieName = "cerca"

type Session struct {
	Store           *sessions.CookieStore
	ShortLivedStore *sessions.CookieStore
}

func New(authKey string, developing bool) *Session {
	store := sessions.NewCookieStore([]byte(authKey))
	store.Options = &sessions.Options{
		HttpOnly: true,
		Secure:   !developing,
		MaxAge:   86400 * 30,
	}
	short := sessions.NewCookieStore([]byte(authKey))
	short.Options = &sessions.Options{
		HttpOnly: true,
		// Secure: true, // TODO (2022-01-05): uncomment when served over https
		MaxAge: 600, // 10 minutes
	}
	return &Session{
		Store:           store,
		ShortLivedStore: short,
	}
}

func (s *Session) Delete(res http.ResponseWriter, req *http.Request) error {
	ed := util.Describe("delete session cookie")
	clearSession := func(store *sessions.CookieStore) error {
		session, err := store.Get(req, cookieName)
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

func getValueFromSession(req *http.Request, store *sessions.CookieStore, key string) (interface{}, error) {
	session, err := store.Get(req, cookieName)
	if err != nil {
		return nil, err
	}
	value, ok := session.Values[key]
	if !ok {
		err := errors.New(fmt.Sprintf("extracting %s from session; no such value", key))
		return nil, util.Eout(err, "get session")
	}
	return value, nil
}

func (s *Session) GetVerificationCode(req *http.Request) (string, error) {
	val, err := getValueFromSession(req, s.ShortLivedStore, "verificationCode")
	if val == nil || err != nil {
		return "", err
	}
	return val.(string), err
}

func (s *Session) Get(req *http.Request) (int, error) {
	val, err := getValueFromSession(req, s.Store, "userid")
	if val == nil || err != nil {
		return -1, err
	}
	return val.(int), err
}

func (s *Session) Save(req *http.Request, res http.ResponseWriter, userid int) error {
	session, _ := s.Store.Get(req, cookieName)
	session.Values["userid"] = userid
	return session.Save(req, res)
}

func (s *Session) SaveVerificationCode(req *http.Request, res http.ResponseWriter, code string) error {
	session, _ := s.ShortLivedStore.Get(req, cookieName)
	session.Values["verificationCode"] = code
	return session.Save(req, res)
}
