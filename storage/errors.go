package storage

import "errors"

var ErrIncorrectUser = errors.New("user doesn't exist or is incorrect")
var ErrNotEnoughPerm = errors.New("not enough permissions")
var ErrTenderNotFound = errors.New("tender wasn't found")
var ErrBidNotFound = errors.New("bid wasn't found")
var ErrVersionNotFound = errors.New("version wasn't found")
var ErrStatusCantBeChanged = errors.New("status cannot be changed")
var ErrTenderClosed = errors.New("tender has already been closed")
