package pkgcommon

const (
	//Inventory Constants
	active          = 1
	soldOut         = 2
	inventoryDomain = "CarMarketPlace"
)

//inventoryConstants
type inventoryConstants struct {
	Domain  string
	Active  uint8
	SoldOut uint8
}

//GetInventoryConstants return inventoryConstants
func GetInventoryConstants() inventoryConstants {
	return inventoryConstants{
		Domain:  inventoryDomain,
		Active:  active,
		SoldOut: soldOut,
	}
}
