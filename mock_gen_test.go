//go:generate mockgen -destination mock_net_test.go -package proxyprotocol_test net Listener,Conn
//go:generate mockgen -destination mock_test.go -package proxyprotocol_test github.com/c0va23/go-proxyprotocol Logger

package proxyprotocol_test
