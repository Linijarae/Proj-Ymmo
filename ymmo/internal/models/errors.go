package models

import "errors"

// ErrForbidden est retourné si l'utilisateur n'a pas les droits
var ErrForbidden = errors.New("accès refusé")

// ErrNotFound est retourné si la ressource n'existe pas
var ErrNotFound = errors.New("ressource introuvable")
