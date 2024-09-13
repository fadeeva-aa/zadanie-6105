package handlers

import "errors"

var ErrPassUsername = errors.New("pass username")
var ErrPassFeedback = errors.New("pass feedback")
var ErrIncorrectStatus = errors.New("incorrect status")
var ErrNothingToDo = errors.New("nothing to do")
