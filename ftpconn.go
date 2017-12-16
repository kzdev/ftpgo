package ftpgo

import "net"

type FtpDataConnector struct {
	conn net.Conn
	c    *Ftp
}

func (r *FtpDataConnector) Read(buf []byte) (int, error) {
	return r.conn.Read(buf)
}

func (r *FtpDataConnector) Write(buf []byte) (int, error) {
	return r.conn.Write(buf)
}

func (r *FtpDataConnector) Close() error {
	err := r.conn.Close()
	_, _, err2 := r.c.getResponse(226)
	if err2 != nil {
		err = err2
	}
	return err
}
