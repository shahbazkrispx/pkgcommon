package pkgcommon

const (
	//Inventory Constants
	active  = 1
	soldOut = 2
)

type inventoryConstants struct {
	Active  uint8
	SoldOut uint8
}

func GetInventoryConstants() inventoryConstants {
	return inventoryConstants{
		Active:  active,
		SoldOut: soldOut,
	}
}
