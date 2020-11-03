package steam

type Filter func(*InventoryItem) bool

func IsTradable(cond bool) Filter {
	return func(item *InventoryItem) bool {
		return (item.Desc.Tradable != 0) == cond
	}
}
