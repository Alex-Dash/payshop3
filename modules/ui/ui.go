package ui

import (
	"payshop3/api"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var PrettyNamesBySKU map[string]string = map[string]string{
	"pd3_preplanning_branchbank_1": "Van Will Not Leave",
	"pd3_preplanning_branchbank_2": "Additional Secure Point",
	"pd3_preplanning_branchbank_3": "Keycard Location",
	"pd3_preplanning_branchbank_4": "Additional Thermite",

	"pd3_preplanning_armoredtransport_1": "Garbage Chute Secure Point",
	"pd3_preplanning_armoredtransport_2": "Unlocked lockboxes",
	"pd3_preplanning_armoredtransport_3": "Additional Explosives",
	"pd3_preplanning_armoredtransport_4": "Stronger Signal",

	"pd3_preplanning_jewelrystore_1": "Distracted Manager",
	"pd3_preplanning_jewelrystore_2": "Escape Van Stays Longer",
	"pd3_preplanning_jewelrystore_3": "Rooftop Chopper",
	"pd3_preplanning_jewelrystore_4": "Employee Entrance",

	"pd3_preplanning_nightclub_1": "Additional Secure Point",
	"pd3_preplanning_nightclub_2": "Vault Code Access",
	"pd3_preplanning_nightclub_3": "Crypto Wallet Timer",
	"pd3_preplanning_nightclub_4": "Inside Man Keycard",

	"pd3_preplanning_artgallery_1": "Dumpster Secure Point",
	"pd3_preplanning_artgallery_2": "Additional QR key",
	"pd3_preplanning_artgallery_3": "Open Main Entrance",
	"pd3_preplanning_artgallery_4": "Faster Chopper",

	"pd3_preplanning_sharkebank_1": "Teller Door",
	"pd3_preplanning_sharkebank_2": "Cafe Celebration",
	"pd3_preplanning_sharkebank_3": "Thermal Lance Parts",
	"pd3_preplanning_sharkebank_4": "Elevator Access",

	"pd3_preplanning_cargodock_1": "Opened Containter",
	"pd3_preplanning_cargodock_2": "Prototype Bags",
	"pd3_preplanning_cargodock_3": "Additional Explosives",
	"pd3_preplanning_cargodock_4": "Thermite Drop",

	"pd3_preplanning_penthouse_1": "Window Cleaning Platform",
	"pd3_preplanning_penthouse_2": "Hidden Thermite",
	"pd3_preplanning_penthouse_3": "Vomiting Agent (x4)",
	"pd3_preplanning_penthouse_4": "Laundry Chute Secure Point",

	"pd3_preplanning_uni_ammobag":  "Ammo Bag",
	"pd3_preplanning_uni_armorbag": "Armor Bag",
	"pd3_preplanning_uni_medicbag": "Medic Bag",
	"pd3_preplanning_uni_zipline":  "Zipline Bag",

	"pd3_coin_goldsmall0":  "1 C-Stack",
	"pd3_coin_goldmedium0": "5 C-Stacks",
	"pd3_coin_goldlarge0":  "10 C-Stacks",
}

var PrettyNamePrefixBySKU map[string]string = map[string]string{
	"uni":              "Universal",
	"branchbank":       "No Rest For The Wicked",
	"armoredtransport": "Road Rage",
	"jewelrystore":     "Dirty Ice",
	"nightclub":        "Rock The Cradle",
	"artgallery":       "Under The Surphaze",
	"sharkebank":       "Gold & Sharke",
	"cargodock":        "99 Boxes",
	"penthouse":        "Touch the Sky",
}

var HeistSelector map[int][]string = map[int][]string{
	0: {"-", "-- SELECT --"},
	1: {"branchbank", "No Rest For The Wicked"},
	2: {"armoredtransport", "Road Rage"},
	3: {"jewelrystore", "Dirty Ice"},
	4: {"nightclub", "Rock The Cradle"},
	5: {"artgallery", "Under The Surphaze"},
	6: {"sharkebank", "Gold & Sharke"},
	7: {"cargodock", "99 Boxes"},
	8: {"penthouse", "Touch the Sky"},
	9: {"all", "EVERYTHING"},
}

var LoaderUIBraile []string = []string{
	"⠻",
	"⠽",
	"⠾",
	"⠷",
	"⠯",
	"⠟",
}

func PrettifyBasic(adg *[]api.AssetGroupData) *[]api.AssetGroupData {
	var ret []api.AssetGroupData = []api.AssetGroupData{}
	for _, v := range *adg {
		pn := PrettyNamePrefixBySKU[v.Sku]
		if pn == "" {
			// No pretty name
			pn = cases.Title(language.English).String(v.Sku)
		}
		v.PrettyName = pn

		item_bank := []api.ShopItemData{}
		for _, x := range v.Bank {
			ipn := PrettyNamesBySKU[*x.Sku]
			if ipn == "" {
				ipn = *x.Name
			}
			x.PrettyName = &ipn
			x.PrettyHeistName = &pn
			item_bank = append(item_bank, x)
		}

		ret = append(ret, api.AssetGroupData{
			GroupName:  cases.Title(language.English).String(v.Sku),
			PrettyName: pn,
			Sku:        v.Sku,
			Bank:       item_bank,
		})

	}
	return &ret
}
