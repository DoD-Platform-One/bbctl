package filesystem

type ClientGetter struct{}

func (clientGetter *ClientGetter) GetClient() Client {
	return NewClient(userHomeDirFunc, createFunc)
}
