package steam

func (c *Client) checkSession() {
	// todo check current session
	for {
		select {
		case <-c.ctx.Done():
			return
		}
	}
}
