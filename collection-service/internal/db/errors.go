package db

import "errors"

var (
    ErrCollectionNotFound     = errors.New("collection not found")
    ErrCollectionItemNotFound = errors.New("collection item not found")
)
