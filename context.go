package authboss

import (
	"context"
	"net/http"
)

type contextKey string

const (
	ctxKeyPID  contextKey = "pid"
	ctxKeyUser contextKey = "user"
)

func (c contextKey) String() string {
	return "authboss ctx key " + string(c)
}

// CurrentUserID retrieves the current user from the session.
func (a *Authboss) CurrentUserID(w http.ResponseWriter, r *http.Request) (string, error) {
	_, err := a.Callbacks.FireBefore(EventGetUserSession, r.Context())
	if err != nil {
		return "", err
	}

	session := a.SessionStoreMaker.Make(w, r)
	key, _ := session.Get(SessionKey)
	return key, nil
}

// CurrentUserIDP retrieves the current user but panics if it's not available for
// any reason.
func (a *Authboss) CurrentUserIDP(w http.ResponseWriter, r *http.Request) string {
	i, err := a.CurrentUserID(w, r)
	if err != nil {
		panic(err)
	} else if len(i) == 0 {
		panic(ErrUserNotFound)
	}

	return i
}

// CurrentUser retrieves the current user from the session and the database.
func (a *Authboss) CurrentUser(w http.ResponseWriter, r *http.Request) (Storer, error) {
	pid, err := a.CurrentUserID(w, r)
	if err != nil {
		return nil, err
	} else if len(pid) == 0 {
		return nil, nil
	}

	return a.currentUser(r.Context(), pid)
}

// CurrentUserP retrieves the current user but panics if it's not available for
// any reason.
func (a *Authboss) CurrentUserP(w http.ResponseWriter, r *http.Request) Storer {
	i, err := a.CurrentUser(w, r)
	if err != nil {
		panic(err)
	}
	return i
}

func (a *Authboss) currentUser(ctx context.Context, pid string) (Storer, error) {
	_, err := a.Callbacks.FireBefore(EventGetUser, ctx)
	if err != nil {
		return nil, err
	}

	user, err := a.StoreLoader.Load(ctx, pid)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, ctxKeyUser, user)
	err = a.Callbacks.FireAfter(EventGetUser, ctx)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// LoadCurrentUser takes a pointer to a pointer to the request in order to
// change the current method's request pointer itself to the new request that
// contains the new context that has the pid in it.
func (a *Authboss) LoadCurrentUserID(w http.ResponseWriter, r **http.Request) (string, error) {
	pid, err := a.CurrentUserID(w, *r)
	if err != nil {
		return "", err
	}

	if len(pid) == 0 {
		return "", nil
	}

	ctx := context.WithValue((**r).Context(), ctxKeyPID, pid)
	*r = (**r).WithContext(ctx)

	return pid, nil
}

func (a *Authboss) LoadCurrentUserIDP(w http.ResponseWriter, r **http.Request) string {
	pid, err := a.LoadCurrentUserID(w, r)
	if err != nil {
		panic(err)
	} else if len(pid) == 0 {
		panic(ErrUserNotFound)
	}

	return pid
}

// LoadCurrentUser takes a pointer to a pointer to the request in order to
// change the current method's request pointer itself to the new request that
// contains the new context that has the user in it. Calls LoadCurrentUserID
// so the primary id is also put in the context.
func (a *Authboss) LoadCurrentUser(w http.ResponseWriter, r **http.Request) (Storer, error) {
	pid, err := a.LoadCurrentUserID(w, r)
	if err != nil {
		return nil, err
	}

	if len(pid) == 0 {
		return nil, nil
	}

	ctx := (**r).Context()
	user, err := a.currentUser(ctx, pid)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, ctxKeyUser, user)
	*r = (**r).WithContext(ctx)
	return user, nil
}

func (a *Authboss) LoadCurrentUserP(w http.ResponseWriter, r **http.Request) Storer {
	user, err := a.LoadCurrentUser(w, r)
	if err != nil {
		panic(err)
	} else if user == nil {
		panic(ErrUserNotFound)
	}

	return user
}
