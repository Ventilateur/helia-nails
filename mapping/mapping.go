package mapping

const (
	employeeTee     = 461942
	employeeJade    = 461945
	employeeChloe   = 461946
	employeeMinette = 461944
)

var (
	allEmployees       = []int{employeeChloe, employeeMinette, employeeJade, employeeTee}
	eyelashesEmployees = []int{employeeChloe}
)

type TreatwellOffer struct {
	OfferID           int
	SkuID             int
	EligibleEmployees []int
}

var ClassPassOfferToTreatWellOffers = map[string][]TreatwellOffer{
	"Beauté des mains avec pose de vernis semi-permanent": {
		{
			OfferID:           4435453,
			SkuID:             8334380,
			EligibleEmployees: allEmployees,
		},
	},
	"Beauté des pieds avec pose de vernis semi-permanent": {
		{
			OfferID:           4435255,
			SkuID:             8403680,
			EligibleEmployees: allEmployees,
		},
	},
	"Beauté des mains et pieds avec pose de vernis semi-permanent": {
		{
			OfferID:           4470648,
			SkuID:             8403747,
			EligibleEmployees: allEmployees,
		},
	},
}
