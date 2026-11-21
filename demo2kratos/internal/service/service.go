package service

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewSvc文章服务)
