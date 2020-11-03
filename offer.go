package steam

type TradeOffer struct {
	ID                 uint64      `json:"tradeofferid,string"`
	Partner            uint32      `json:"accountid_other"`
	ReceiptID          uint64      `json:"tradeid,string"`
	RecvItems          []*EconItem `json:"items_to_receive"`
	SendItems          []*EconItem `json:"items_to_give"`
	Message            string      `json:"message"`
	State              uint8       `json:"trade_offer_state"`
	ConfirmationMethod uint8       `json:"confirmation_method"`
	Created            int64       `json:"time_created"`
	Updated            int64       `json:"time_updated"`
	Expires            int64       `json:"expiration_time"`
	EscrowEndDate      int64       `json:"escrow_end_date"`
	RealTime           bool        `json:"from_real_time_trade"`
	IsOurOffer         bool        `json:"is_our_offer"`
}

func (offer *TradeOffer) Send(c *Client, sid SteamID, token string) error {
	return c.SendTradeOffer(offer, sid, token)
}

func (offer *TradeOffer) Accept(c *Client) error {
	return c.AcceptTradeOffer(offer.ID)
}

func (offer *TradeOffer) Cancel(c *Client) error {
	if offer.IsOurOffer {
		return c.CancelTradeOffer(offer.ID)
	}

	return c.DeclineTradeOffer(offer.ID)
}
