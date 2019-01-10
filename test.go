func handleTcpConnection(conn net.Conn, w *Worker) {
	buffer, err := encodeBuffer(<-buildQueue)
	if _, err := conn.Write(buffer.Bytes());  err != nil {
		log.Print(err)
	}
	tmpBuff := make([]byte, 1000)
	if _, err := conn.Read(tmpBuff); err != nil {
		log.Print(err)
	}
}
