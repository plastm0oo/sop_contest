package service

import "errors"

// чтобы не дублировать в хендлере и юзкейсе
var ErrEmptyMessage = errors.New("message can't be empty")
