package pkgcommon

const (
	//Inventory Constants
	deActive        = 0
	active          = 1
	soldOut         = 2
	pending         = 3
	carDeleted      = "CarDeleted"
	carSoldOut      = "CarSoldOut"
	carUpdated      = "CarUpdated"
	inventoryDomain = "CarMarketPlace"
)

//inventoryConstants
//all constants related to inventory
type inventoryConstants struct {
	Domain     string
	DeActive   uint8
	Active     uint8
	SoldOut    uint8
	Pending    uint8
	CarSoldOut string
	CarUpdated string
	CarDeleted string
}

//GetInventoryConstants return inventoryConstants
func GetInventoryConstants() inventoryConstants {
	return inventoryConstants{
		Domain:     inventoryDomain,
		DeActive:   deActive,
		Active:     active,
		SoldOut:    soldOut,
		Pending:    pending,
		CarDeleted: carDeleted,
		CarSoldOut: carSoldOut,
		CarUpdated: carUpdated,
	}
}
