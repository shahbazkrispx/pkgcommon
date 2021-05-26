package pkgcommon

const (
	//Inventory Constants
	active  = 1
	soldOut = 2
)

//inventoryConstants
type inventoryConstants struct {
	Active  uint8
	SoldOut uint8
}

//GetInventoryConstants return inventoryConstants
func GetInventoryConstants() inventoryConstants {
	return inventoryConstants{
		Active:  active,
		SoldOut: soldOut,
	}
}
