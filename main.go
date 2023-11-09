/*
PayShop3 - An Interactive Order-based System for PayDay3
Source: https://github.com/Alex-Dash/payshop3
Copyright (C) 2023  AlexDash
*/
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"payshop3/api"
	"payshop3/ui"
	"payshop3/util"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	loginScreen     *tview.Grid
	loginForm       *tview.Form
	entryPage       *tview.Grid
	main_menu_list  *tview.List
	app             *tview.Application
	pages           *tview.Pages
	UI_header_info  *tview.Grid
	cart_section    *tview.Grid
	basicOrderData  api.BasicOrderData
	exOrderData     api.ExclusiveOrderData
	goldOrderData   api.GoldOrderData
	credOrderData   api.CreditOrderData
	Cart            []api.OrderInitData
	checkout        func()
	OrderInProgress bool
	B_VER           = "v0.8.5-ALPHA"
)

func main() {
	// Keep the app running
	// sc := make(chan os.Signal, 1)
	// signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	login_raw, err := os.ReadFile("payshop3_logindata.json")
	if err == nil {
		ta := tview.NewApplication()
		p := newPrimitive("Logging you in, please wait...")
		go func() {
			if err := ta.SetRoot(p, true).Run(); err != nil {
				panic(err)
			}
		}()

		var d api.LoginData
		err = json.Unmarshal(login_raw, &d)
		if err == nil {
			api.Init(d.Login, d.Password, d.AutoLogin)
		}
		ta.Stop()
	}
	setupUI()

	// <-sc
	fmt.Println("\n=======================\nQuitting PayShop3...\n=======================\n")
}

func onlyNumbers(s string, r rune) bool {
	_, err := strconv.Atoi(s + string(r))
	return err == nil
}

func formatNumberSpaced(n int) string {
	r := ""
	a := strings.Split(strconv.Itoa(n), "")
	t := 0
	for i := len(a) - 1; i >= 0; i-- {
		if t > 0 && t%3 == 0 {
			r = " " + r
		}
		t += 1
		r = a[i] + r
	}
	return r
}

func genericModal(text string) {
	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{"Back"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.SetRoot(pages, true).SetFocus(pages)
		})
	app.SetRoot(modal, true).SetFocus(modal)
}

func browserModal(resp api.OrderRespData) {
	dec := *resp.Currency.Decimals
	curr := *resp.Currency.CurrencyCode
	sym := ui.CurrencySumbolByCode[curr]

	price := fmt.Sprintf("%s%s %s", sym, util.ToFixedDecimal(*resp.Price, dec), curr)
	tax := fmt.Sprintf("%s%s %s", sym, util.ToFixedDecimal(*resp.Tax, dec), curr)
	vat := fmt.Sprintf("%s%s %s", sym, util.ToFixedDecimal(*resp.Vat, dec), curr)
	stax := fmt.Sprintf("%s%s %s", sym, util.ToFixedDecimal(*resp.SalesTax, dec), curr)
	ppfee := fmt.Sprintf("%s%s %s", sym, util.ToFixedDecimal(*resp.PaymentProviderFee, dec), curr)
	pmfee := fmt.Sprintf("%s%s %s", sym, util.ToFixedDecimal(*resp.PaymentMethodFee, dec), curr)

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Your order has been placed\nOrder No: %s\nSubtotal: %s\nTax: %s\nVAT: %s\nSales Tax: %s\nPayment Provider Fee: %s\nPayment Method Fee: %s\nLink: %s\n",
			*resp.OrderNo, price, tax, vat, stax, ppfee, pmfee, *resp.PaymentStationUrl)).
		AddButtons([]string{"Back", "Open in browser"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Open in browser" {
				util.OpenBrowser(*resp.PaymentStationUrl)
			}
			app.SetRoot(pages, true).SetFocus(pages)
		})
	app.SetRoot(modal, true).SetFocus(modal)
}

func updateHeaderUI() {
	if UI_header_info != nil {
		entryPage.RemoveItem(UI_header_info)
	}
	logout := tview.NewButton("Log out").
		SetSelectedFunc(func() {
			api.Logout()
			pages.SwitchToPage("login")
		})

	cash, err1 := api.GetCachedWalletByCode("CASH")
	gold, err2 := api.GetCachedWalletByCode("GOLD")
	cred, err3 := api.GetCachedWalletByCode("CRED")

	if api.LD.DisplayName != "" && err1 == nil && err2 == nil && err3 == nil {
		UI_header_info = tview.NewGrid().SetRows(1).SetColumns(0, 40, 10).
			AddItem(newPrimitive(fmt.Sprintf("Cash: $%s | C-Stacks: %s | Credits: %s",
				formatNumberSpaced(*cash.Balance),
				formatNumberSpaced(*gold.Balance),
				formatNumberSpaced(*cred.Balance))), 0, 0, 1, 1, 0, 0, false).
			AddItem(newPrimitive(fmt.Sprintf("LOGGED IN AS: %s", api.LD.DisplayName)), 0, 1, 1, 1, 0, 0, false).
			AddItem(logout, 0, 2, 1, 1, 0, 0, false)
	} else {
		UI_header_info = tview.NewGrid().SetRows(1).SetColumns(0, 40, 10).
			AddItem(newPrimitive("Cash: $999 999 999 999 | C-Stacks:99 999 | Credits: 999 999"), 0, 0, 1, 1, 0, 0, false).
			AddItem(newPrimitive("Logged in as: USERNAME_USERNAME"), 0, 1, 1, 1, 0, 0, false).
			AddItem(logout, 0, 2, 1, 1, 0, 0, false)
	}

	entryPage.AddItem(UI_header_info, 0, 0, 1, 3, 0, 0, false)
	time.AfterFunc(time.Second*2, func() { app.Draw() })
}

func newPrimitive(text string) tview.Primitive {
	return tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(text)
}

func addBasicCacheToCart() error {
	if OrderInProgress {
		return nil
	}
	if basicOrderData.ItemTypeID == 0 {
		return errors.New("you have to specify the Item Type")
	}
	if basicOrderData.BuyTypeID == 0 {
		return errors.New("you have to specify the Buy Type")
	}

	if basicOrderData.Amount == 0 {
		return errors.New("you cannot place an order for 0 items")
	}

	// data is valid, proceed
	rawCache := api.GetAssetBank()
	assetCache := ui.PrettifyBasic(&rawCache)
	var itemRef []api.ShopItemData

	for _, sel1 := range *assetCache {
		if sel1.Sku == "uni" {
			for _, v := range sel1.Bank {
				if strings.Contains(strings.ToLower(*v.Name), strings.ToLower(basicOrderData.ItemType)) || (basicOrderData.ItemTypeID == 5) {
					itemRef = append(itemRef, v)
				}
			}
		}
	}
	if len(itemRef) == 0 {
		return errors.New("could not find item in the shop")
	}

	// item ref valid, proceed

	var count int
	switch basicOrderData.BuyTypeID {
	case 1:
		count = basicOrderData.Amount
	case 2:
		bundle_discount_price := 0
		for _, lref := range itemRef {
			rd := *lref.RegionData
			bundle_discount_price += *rd[0].DiscountedPrice
		}
		if bundle_discount_price == 0 {
			count = 1
		} else {
			count = basicOrderData.Amount / bundle_discount_price
		}

	default:
		return errors.New("failed to guess buy order type")
	}

	for _, lref := range itemRef {
		if count == 0 {
			// do not add 0 quantity items
			continue
		}
		rd := *lref.RegionData
		orderObj := api.OrderInitData{
			ItemId:          *lref.ItemId,
			Quantity:        count,
			Price:           *rd[0].Price * count,
			DiscountedPrice: *rd[0].DiscountedPrice * count,
			CurrencyCode:    *rd[0].CurrencyCode,
			Region:          *lref.Region,
			Language:        *lref.Language,
			PrettyName:      *lref.PrettyName,
			PrettyHeistName: *lref.PrettyHeistName,
			ReturnUrl:       "http://127.0.0.1",
		}
		Cart = append(Cart, orderObj)
	}

	//order cached in cart, update UI
	updateCartUI()

	return nil
}

func addExclusiveCacheToCart() error {
	if OrderInProgress {
		return nil
	}
	if exOrderData.BuyTypeID == 0 {
		return errors.New("you have to specify the Buy Type")
	}

	if exOrderData.HeistTypeID == 0 {
		return errors.New("you have to select the hesit first")
	}

	if exOrderData.HeistType != "EVERYTHING" && exOrderData.ItemTypeID == 0 {
		return errors.New("asset was not selected")
	}

	if exOrderData.HeistType == "EVERYTHING" && exOrderData.ItemType != "EVERYTHING" {
		return errors.New("asset selection is incorrect")
	}

	rawCache := api.GetAssetBank()
	assetCache := ui.PrettifyBasic(&rawCache)
	var itemRef []api.ShopItemData

	g_sku := ui.HeistSelector[exOrderData.HeistTypeID][0]

	for _, sel1 := range *assetCache {
		if (sel1.Sku == g_sku || exOrderData.HeistType == "EVERYTHING") && sel1.Sku != "uni" {
			for _, v := range sel1.Bank {
				if exOrderData.HeistType == "EVERYTHING" {
					itemRef = append(itemRef, v)
					continue
				}
				if *v.Sku == exOrderData.ItemTypeSKU || (exOrderData.ItemType == "EVERYTHING") {
					itemRef = append(itemRef, v)
				}
			}
		}
	}
	if len(itemRef) == 0 {
		return errors.New("could not find item in the shop")
	}

	// item ref valid, proceed

	var count int
	switch exOrderData.BuyTypeID {
	case 1:
		count = exOrderData.Amount
	case 2:
		bundle_discount_price := 0
		for _, lref := range itemRef {
			rd := *lref.RegionData
			bundle_discount_price += *rd[0].DiscountedPrice
		}
		if bundle_discount_price == 0 {
			count = 1
		} else {
			count = exOrderData.Amount / bundle_discount_price
		}

	default:
		return errors.New("failed to guess buy order type")
	}

	for _, lref := range itemRef {
		if count == 0 {
			// do not add 0 quantity items
			continue
		}
		rd := *lref.RegionData
		orderObj := api.OrderInitData{
			ItemId:          *lref.ItemId,
			Quantity:        count,
			Price:           *rd[0].Price * count,
			DiscountedPrice: *rd[0].DiscountedPrice * count,
			CurrencyCode:    *rd[0].CurrencyCode,
			Region:          *lref.Region,
			Language:        *lref.Language,
			PrettyName:      *lref.PrettyName,
			PrettyHeistName: *lref.PrettyHeistName,
			ReturnUrl:       "http://127.0.0.1",
		}
		Cart = append(Cart, orderObj)
	}

	//order cached in cart, update UI
	updateCartUI()

	return nil
}

func addGoldCacheToCart() error {
	if OrderInProgress {
		return nil
	}
	if goldOrderData.BuyTypeID == 0 {
		return errors.New("you have to specify the Buy Type")
	}

	if goldOrderData.Amount == 0 {
		return errors.New("cannot buy 0 C-Stacks")
	}

	g1, err := api.GetItemBySKU("pd3_coin_goldsmall0")
	if err != nil {
		return errors.New("could not find 1 C-Stack bundle in the shop")
	}

	g5, err := api.GetItemBySKU("pd3_coin_goldmedium0")
	if err != nil {
		return errors.New("could not find 5 C-Stack bundle in the shop")
	}

	g10, err := api.GetItemBySKU("pd3_coin_goldlarge0")
	if err != nil {
		return errors.New("could not find 10 C-Stack bundle in the shop")
	}
	presum := goldOrderData.Amount
	switch goldOrderData.BuyTypeID {
	case 1:
		// by coin amount
		for _, g := range []api.ShopItemData{g10, g5, g1} {
			count := presum / *g.UseCount
			if count == 0 {
				continue
			}
			rd := *g.RegionData
			Cart = append(Cart, api.OrderInitData{
				ItemId:          *g.ItemId,
				Quantity:        count,
				Price:           *rd[0].Price * count,
				DiscountedPrice: *rd[0].DiscountedPrice * count,
				CurrencyCode:    *rd[0].CurrencyCode,
				Region:          *g.Region,
				Language:        *g.Language,
				ReturnUrl:       "http://127.0.0.1/",
				PrettyName:      ui.PrettyNamesBySKU[*g.Sku],
				PrettyHeistName: "Universal",
			})
			presum = presum - (*g.UseCount * count)
		}
	case 2:
		// by wallet amount
		for _, g := range []api.ShopItemData{g10, g5, g1} {
			rd := *g.RegionData
			count := presum / *rd[0].DiscountedPrice
			if count == 0 {
				continue
			}
			Cart = append(Cart, api.OrderInitData{
				ItemId:          *g.ItemId,
				Quantity:        count,
				Price:           *rd[0].Price * count,
				DiscountedPrice: *rd[0].DiscountedPrice * count,
				CurrencyCode:    *rd[0].CurrencyCode,
				Region:          *g.Region,
				Language:        *g.Language,
				ReturnUrl:       "http://127.0.0.1/",
				PrettyName:      ui.PrettyNamesBySKU[*g.Sku],
				PrettyHeistName: "Universal",
			})
			presum = presum - (*rd[0].DiscountedPrice * count)
		}

	default:
		return errors.New("unacceptable order type")
	}
	updateCartUI()

	return nil
}

func orderCredits(credit_shop_items []api.ShopItemData, form *tview.Form) (api.OrderRespData, error) {
	b := form.GetButton(form.GetButtonIndex("Order directly"))
	if b == nil {
		return api.OrderRespData{}, errors.New("failed to find ui button")
	}
	b.SetDisabled(true)
	f := false
	var item api.ShopItemData
	for _, item = range credit_shop_items {
		if *item.Name == credOrderData.ItemType {
			f = true
			break
		}
	}
	if !f {
		b.SetDisabled(false)
		return api.OrderRespData{}, errors.New("could not find item in the shop")
	}
	rd := *item.RegionData
	oid := api.OrderInitData{
		ItemId:          *item.ItemId,
		Quantity:        credOrderData.Amount,
		Price:           *rd[0].Price * credOrderData.Amount,
		DiscountedPrice: *rd[0].DiscountedPrice * credOrderData.Amount,
		CurrencyCode:    *rd[0].CurrencyCode,
		Region:          *item.Region,
		Language:        *item.Language,
		ReturnUrl:       "http://127.0.0.1",
	}
	resp, err := api.ExecOrder(oid)
	if resp.PaymentStationUrl != nil && err == nil {
		// link present
		b.SetDisabled(false)
		return resp, nil
	}
	b.SetDisabled(false)
	if err != nil {
		return api.OrderRespData{}, err
	}
	return api.OrderRespData{}, errors.New("failed to find payment link")
}

func headerTimedUpdate() {
	api.UpdateWallets()
	updateHeaderUI()
	time.AfterFunc(time.Minute, headerTimedUpdate)
}

func updateCartUI() {
	if cart_section != nil {
		entryPage.RemoveItem(cart_section)
	}
	cart_table := tview.NewTable().SetBorders(true)
	for c, v := range []string{"#", "Name", "Price", "Qty", "Subtotal", "Currency", "DEL"} {
		cart_table.SetCell(0, c, tview.NewTableCell(v).SetAlign(tview.AlignCenter).SetTextColor(tcell.ColorYellow))
	}

	total_tbl := tview.NewTable().SetBorders(true)
	for c, v := range []string{"Currency", "Subtotal", "Discounted %", "Total"} {
		total_tbl.SetCell(0, c, tview.NewTableCell(v).SetAlign(tview.AlignCenter).SetTextColor(tcell.ColorYellow))
	}

	totalmap := make(map[string][]int)

	for i, v := range Cart {
		cc := v.CurrencyCode
		if v.CurrencyCode == "GOLD" {
			cc = "C-STACKS"
		}
		cart_table.SetCell(i+1, 0, tview.NewTableCell(formatNumberSpaced(i+1)).SetAlign(tview.AlignLeft))
		cart_table.SetCell(i+1, 1, tview.NewTableCell(v.PrettyName).SetAlign(tview.AlignLeft))
		cart_table.SetCell(i+1, 2, tview.NewTableCell(formatNumberSpaced(v.Price/v.Quantity)).SetAlign(tview.AlignLeft))
		cart_table.SetCell(i+1, 3, tview.NewTableCell(formatNumberSpaced(v.Quantity)).SetAlign(tview.AlignLeft))
		cart_table.SetCell(i+1, 4, tview.NewTableCell(formatNumberSpaced(v.Price)).SetAlign(tview.AlignLeft))
		cart_table.SetCell(i+1, 5, tview.NewTableCell(cc).SetAlign(tview.AlignLeft))
		cart_table.SetCell(i+1, 6, tview.NewTableCell(" |X| ").SetAlign(tview.AlignCenter).SetTextColor(tcell.ColorRed))
		cart_table.SetSelectionChangedFunc(func(row, column int) {
			if column == cart_table.GetColumnCount()-1 && row > 0 {
				ct := []api.OrderInitData{}
				for i, v := range Cart {
					if i+1 == row {
						continue
					}
					ct = append(ct, v)
				}
				Cart = ct
				updateCartUI()
			}
		}).SetSelectable(true, false)
		if len(totalmap[cc]) != 0 {
			// key exists. Update values
			totalmap[cc] = []int{totalmap[cc][0] + v.Price, totalmap[cc][1] + v.DiscountedPrice}
		} else {
			totalmap[cc] = []int{v.Price, v.DiscountedPrice}
		}
	}

	offset := 0
	for k, v := range totalmap {
		dp := (1 - float64(v[1])/float64(v[0])) * 10000
		dp = math.Round(dp) / 100
		total_tbl.SetCell(offset+1, 0, tview.NewTableCell(k).SetAlign(tview.AlignLeft))
		total_tbl.SetCell(offset+1, 1, tview.NewTableCell(formatNumberSpaced(v[0])).SetAlign(tview.AlignLeft))
		total_tbl.SetCell(offset+1, 2, tview.NewTableCell(fmt.Sprintf("%v", dp)+"%").SetAlign(tview.AlignLeft))
		total_tbl.SetCell(offset+1, 3, tview.NewTableCell(formatNumberSpaced(v[1])).SetAlign(tview.AlignLeft))
		offset++
	}

	clr_cart_btn := tview.NewButton("Clear All").SetStyle(tcell.Style{}.Background(tcell.ColorDarkRed)).
		SetSelectedFunc(func() {
			Cart = []api.OrderInitData{}
			updateCartUI()
		})

	cart_top := tview.NewGrid().SetColumns(0, 10).
		AddItem(newPrimitive("Your Cart"), 0, 0, 1, 1, 0, 0, false).
		AddItem(clr_cart_btn, 0, 1, 1, 1, 0, 0, false)

	cart_bottom := tview.NewGrid().
		AddItem(tview.NewButton("Proceed To Checkout").SetSelectedFunc(checkout), 0, 1, 1, 1, 0, 0, false)

	cart_section = tview.NewGrid().SetRows(1, 0, 10, 1).
		AddItem(cart_top, 0, 0, 1, 1, 0, 0, false).
		AddItem(cart_table, 1, 0, 1, 1, 0, 0, false).
		AddItem(total_tbl, 2, 0, 1, 1, 2, 0, false).
		AddItem(cart_bottom, 3, 0, 1, 1, 0, 0, false)

	entryPage.AddItem(cart_section, 1, 2, 1, 1, 0, 130, false)
}

func setupUI() {
	app = tview.NewApplication()
	pages = tview.NewPages()
	loginForm = tview.NewForm().
		AddTextView("Status", "Logged out.\nPlease log in with your Nebula account first", 50, 2, true, false).
		AddInputField("Login", "", 50, nil, func(text string) {}).
		AddPasswordField("Password", "", 50, '*', nil).
		AddCheckbox("Save my info", false, nil).
		AddButton("Login", func() {
			loginForm.GetFormItemByLabel("Status").(*tview.TextView).SetText("Logging in...")
			login := loginForm.GetFormItemByLabel("Login").(*tview.InputField).GetText()
			password := loginForm.GetFormItemByLabel("Password").(*tview.InputField).GetText()
			save := loginForm.GetFormItemByLabel("Save my info").(*tview.Checkbox).IsChecked()
			err := api.Init(login, password, save)
			if err != nil {
				loginForm.GetFormItemByLabel("Status").(*tview.TextView).SetText("Error: " + err.Error())
				return
			}
			loginForm.GetFormItemByLabel("Status").(*tview.TextView).SetText("Loading shop data...")
			err = api.UpdateShop()
			if err != nil {
				loginForm.GetFormItemByLabel("Status").(*tview.TextView).SetText("Error: Could not load shop data. Cannot proceed.")
				return
			}
			pages.SwitchToPage("entry")
			// clear data
			loginForm.GetFormItemByLabel("Status").(*tview.TextView).SetText("Logged out.\nPlease log in with your Nebula account first")
			loginForm.GetFormItemByLabel("Login").(*tview.InputField).SetText("")
			loginForm.GetFormItemByLabel("Password").(*tview.InputField).SetText("")
			loginForm.GetFormItemByLabel("Save my info").(*tview.Checkbox).SetChecked(false)
			updateHeaderUI()
		}).
		AddButton("Quit", func() {
			app.Stop()
		}).SetButtonsAlign(tview.AlignCenter)
	loginScreen = tview.NewGrid().SetColumns(0, 80, 0).SetRows(0, 80, 0).AddItem(loginForm, 1, 1, 1, 1, 0, 0, false)

	// menu := newPrimitive("Menu")
	order_config_basic := newPrimitive("Order configuration")

	cart_section = tview.NewGrid().AddItem(newPrimitive("Your Cart"), 0, 0, 1, 1, 1, 0, false)

	var order_form *tview.Form
	basic_sel := func() {
		if order_form != nil {
			entryPage.RemoveItem(order_form)
		}
		order_form = tview.NewForm().
			AddDropDown("Item Type", []string{"-- SELECT --", "Ammo Bag", "Armor Bag", "Medic Bag", "Zipline Bag", "EVERYTHING"}, 0, func(option string, optionIndex int) {
				basicOrderData.ItemType = option
				basicOrderData.ItemTypeID = optionIndex
			}).
			AddDropDown("Buy order type", []string{"-- SELECT --", "By Item Amount", "By Wallet Amount Limit"}, 0, func(option string, optionIndex int) {
				basicOrderData.BuyType = option
				basicOrderData.BuyTypeID = optionIndex
			}).
			AddInputField("Amount", "", 20, onlyNumbers, func(text string) {
				n, err := strconv.Atoi(text)
				if err == nil {
					basicOrderData.Amount = n
				}
			}).
			AddButton("Cancel", func() {
				entryPage.RemoveItem(order_form).AddItem(order_config_basic, 1, 1, 1, 1, 0, 100, false)
				app.SetFocus(main_menu_list)
			}).
			AddButton("Add to cart", func() {
				err := addBasicCacheToCart()
				if err != nil {
					genericModal(fmt.Sprintf("Error: %s", err.Error()))
				}
			})
		order_form.SetBorder(true).SetTitle("Order configuration").SetTitleAlign(tview.AlignCenter)
		entryPage.RemoveItem(order_config_basic).AddItem(order_form, 1, 1, 1, 1, 0, 100, false)
		app.SetFocus(order_form)
	}

	// exclusive asset thingy
	exclusive_sel := func() {
		if order_form != nil {
			entryPage.RemoveItem(order_form)
		}
		// get all pretty names for selectors
		sel_1 := make([]string, 0, len(ui.HeistSelector))
		for i := 0; i < len(ui.HeistSelector); i++ {
			sel_1 = append(sel_1, ui.HeistSelector[i][1])
		}

		sel2 := []string{"-- SELECT --"}
		order_form = tview.NewForm().
			AddDropDown("Heist", sel_1, exOrderData.HeistTypeID, func(option string, optionIndex int) {
				exOrderData.HeistType = option
				exOrderData.HeistTypeID = optionIndex
				exOrderData.ItemTypeSKU = "-"
				if order_form != nil && order_form.GetFormItemByLabel("Asset") != nil {
					if option == "EVERYTHING" {
						order_form.GetFormItemByLabel("Asset").(*tview.DropDown).SetOptions([]string{"EVERYTHING"}, func(text string, index int) {
							exOrderData.ItemType = text
							exOrderData.ItemTypeID = index
							exOrderData.ItemTypeSKU = "-"
						}).SetCurrentOption(0)
						return
					}
					// get sku of selection
					sel2 = []string{"-- SELECT --"}
					group_sku := ui.HeistSelector[exOrderData.HeistTypeID][0]

					// asset bank
					ab_raw := []api.AssetGroupData{api.GetExclusiveAssetGroupBySku(group_sku)}
					ab := (*ui.PrettifyBasic(&ab_raw))[0]
					i_sku_map := make(map[string]string)
					for _, asset := range ab.Bank {
						sel2 = append(sel2, *asset.PrettyName)
						i_sku_map[*asset.PrettyName] = *asset.Sku
					}
					if len(ab.Bank) > 0 {
						sel2 = append(sel2, "EVERYTHING")
					}

					order_form.GetFormItemByLabel("Asset").(*tview.DropDown).SetOptions(sel2, func(text string, index int) {
						exOrderData.ItemType = text
						exOrderData.ItemTypeID = index
						exOrderData.ItemTypeSKU = i_sku_map[text]
					}).SetCurrentOption(0)
				}
			}).
			AddDropDown("Asset", sel2, 0, func(option string, optionIndex int) {
				exOrderData.ItemType = option
				exOrderData.ItemTypeID = optionIndex
				exOrderData.ItemTypeSKU = "-"
			}).
			AddDropDown("Buy order type", []string{"-- SELECT --", "By Item Amount", "By Wallet Amount Limit"}, 0, func(option string, optionIndex int) {
				exOrderData.BuyType = option
				exOrderData.BuyTypeID = optionIndex
			}).
			AddInputField("Amount", "", 20, onlyNumbers, func(text string) {
				n, err := strconv.Atoi(text)
				if err == nil {
					exOrderData.Amount = n
				}
			}).
			AddButton("Cancel", func() {
				entryPage.RemoveItem(order_form).AddItem(order_config_basic, 1, 1, 1, 1, 0, 100, false)
				app.SetFocus(main_menu_list)
			}).
			AddButton("Add to cart", func() {
				err := addExclusiveCacheToCart()
				if err != nil {
					genericModal(fmt.Sprintf("Error: %s", err.Error()))
				}
			})
		order_form.SetBorder(true).SetTitle("Order configuration").SetTitleAlign(tview.AlignCenter)
		entryPage.RemoveItem(order_config_basic).AddItem(order_form, 1, 1, 1, 1, 0, 100, false)
		app.SetFocus(order_form)
	}

	// C-Stacks market
	gold_sel := func() {
		if order_form != nil {
			entryPage.RemoveItem(order_form)
		}
		order_form = tview.NewForm().
			AddDropDown("Buy order type", []string{"-- SELECT --", "By Coin Amount", "By Wallet Amount Limit"}, 0, func(option string, optionIndex int) {
				goldOrderData.BuyType = option
				goldOrderData.BuyTypeID = optionIndex
			}).
			AddInputField("Amount", "", 20, onlyNumbers, func(text string) {
				n, err := strconv.Atoi(text)
				if err == nil {
					goldOrderData.Amount = n
				}
			}).
			AddButton("Cancel", func() {
				entryPage.RemoveItem(order_form).AddItem(order_config_basic, 1, 1, 1, 1, 0, 100, false)
				app.SetFocus(main_menu_list)
			}).
			AddButton("Add to cart", func() {
				err := addGoldCacheToCart()
				if err != nil {
					genericModal(fmt.Sprintf("Error: %s", err.Error()))
				}
			})
		order_form.SetBorder(true).SetTitle("Order configuration").SetTitleAlign(tview.AlignCenter)
		entryPage.RemoveItem(order_config_basic).AddItem(order_form, 1, 1, 1, 1, 0, 100, false)
		app.SetFocus(order_form)
	}

	checkout = func() {
		if cart_section != nil {
			entryPage.RemoveItem(cart_section)
		}
		checkout_table := tview.NewTable().SetBorders(true)
		for c, v := range []string{"#", "Name", "Price", "Qty", "Subtotal", "Currency", "Status"} {
			checkout_table.SetCell(0, c, tview.NewTableCell(v).SetAlign(tview.AlignCenter).SetTextColor(tcell.ColorYellow))
		}

		total_tbl := tview.NewTable().SetBorders(true)
		for c, v := range []string{"Currency", "Subtotal", "Discounted %", "Total"} {
			total_tbl.SetCell(0, c, tview.NewTableCell(v).SetAlign(tview.AlignCenter).SetTextColor(tcell.ColorYellow))
		}

		totalmap := make(map[string][]int)

		for i, v := range Cart {
			cc := v.CurrencyCode
			if v.CurrencyCode == "GOLD" {
				cc = "C-STACKS"
			}
			checkout_table.SetCell(i+1, 0, tview.NewTableCell(formatNumberSpaced(i+1)).SetAlign(tview.AlignLeft))
			checkout_table.SetCell(i+1, 1, tview.NewTableCell(v.PrettyName).SetAlign(tview.AlignLeft))
			checkout_table.SetCell(i+1, 2, tview.NewTableCell(formatNumberSpaced(v.Price/v.Quantity)).SetAlign(tview.AlignLeft))
			checkout_table.SetCell(i+1, 3, tview.NewTableCell(formatNumberSpaced(v.Quantity)).SetAlign(tview.AlignLeft))
			checkout_table.SetCell(i+1, 4, tview.NewTableCell(formatNumberSpaced(v.Price)).SetAlign(tview.AlignLeft))
			checkout_table.SetCell(i+1, 5, tview.NewTableCell(cc).SetAlign(tview.AlignLeft))
			checkout_table.SetCell(i+1, 6, tview.NewTableCell(" - ").SetAlign(tview.AlignCenter).SetTextColor(tcell.ColorRed)).SetSelectable(true, false)
			if len(totalmap[cc]) != 0 {
				// key exists. Update values
				totalmap[cc] = []int{totalmap[cc][0] + v.Price, totalmap[cc][1] + v.DiscountedPrice}
			} else {
				totalmap[cc] = []int{v.Price, v.DiscountedPrice}
			}
		}

		offset := 0
		for k, v := range totalmap {
			dp := (1 - float64(v[1])/float64(v[0])) * 10000
			dp = math.Round(dp) / 100
			total_tbl.SetCell(offset+1, 0, tview.NewTableCell(k).SetAlign(tview.AlignLeft))
			total_tbl.SetCell(offset+1, 1, tview.NewTableCell(formatNumberSpaced(v[0])).SetAlign(tview.AlignLeft))
			total_tbl.SetCell(offset+1, 2, tview.NewTableCell(fmt.Sprintf("%v", dp)+"%").SetAlign(tview.AlignLeft))
			total_tbl.SetCell(offset+1, 3, tview.NewTableCell(formatNumberSpaced(v[1])).SetAlign(tview.AlignLeft))
			offset++
		}

		order_top := tview.NewGrid().SetColumns(0, 10).
			AddItem(newPrimitive("Your Order"), 0, 0, 1, 1, 0, 0, false)

		var (
			back_btn *tview.Button
			stop_btn *tview.Button
			exec_btn *tview.Button
		)

		back_btn = tview.NewButton("Back To Cart").SetSelectedFunc(func() {
			if OrderInProgress {
				return
			}
			updateCartUI()
		})
		stop_btn = tview.NewButton("Stop Order").SetSelectedFunc(func() {
			if !OrderInProgress {
				return
			}
			back_btn.SetDisabled(false)
			stop_btn.SetDisabled(true)
			exec_btn.SetDisabled(false)
			OrderInProgress = false
		})
		exec_btn = tview.NewButton("Execute Order").SetSelectedFunc(func() {
			go func() {
				if OrderInProgress {
					return
				}
				back_btn.SetDisabled(true)
				stop_btn.SetDisabled(false)
				exec_btn.SetDisabled(true)
				OrderInProgress = true
				for i, cart_item := range Cart {
					if !OrderInProgress {
						return
					}
					var err_p *error

					go func() {
						f := 0
						for {
							if err_p != nil {
								return
							}
							checkout_table.SetCell(i+1, 6, tview.NewTableCell(ui.LoaderUIBraile[f%len(ui.LoaderUIBraile)]).
								SetTextColor(tcell.ColorOrange).
								SetAlign(tview.AlignCenter))
							app.Draw()
							time.Sleep(time.Millisecond * 100)
							f++
						}
					}()
					time.Sleep(time.Millisecond * 1500) // Throttle requests
					od, err := api.ExecOrder(cart_item)

					// return
					err_p = &err
					if err == nil {
						if *od.Status == "FULFILLED" {
							checkout_table.SetCell(i+1, 6, tview.NewTableCell(" ✓ ").SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignCenter))
						} else {
							checkout_table.SetCell(i+1, 6, tview.NewTableCell(" ! ").SetTextColor(tcell.ColorDarkOrange).SetAlign(tview.AlignCenter))
						}
					} else {
						checkout_table.SetCell(i+1, 6, tview.NewTableCell(" X ").SetTextColor(tcell.ColorRed).SetAlign(tview.AlignCenter))
					}
					app.Draw()
				}
				// Order finished
				OrderInProgress = false
				back_btn.SetDisabled(false)
				stop_btn.SetDisabled(true)
				exec_btn.SetDisabled(false)
				api.UpdateWallets()
				updateHeaderUI()
				// show popup
				genericModal("Order has been finished\nPlease restart your game to see your new assets")
				app.Draw()
			}()
		})

		back_btn.SetDisabled(false)
		stop_btn.SetDisabled(true)
		exec_btn.SetDisabled(false)

		// button styles
		back_btn.SetDisabledStyle(tcell.Style{}.Background(tcell.ColorDarkGray).Foreground(tcell.ColorGray))
		stop_btn.SetDisabledStyle(tcell.Style{}.Background(tcell.Color88).Foreground(tcell.Color245))
		exec_btn.SetDisabledStyle(tcell.Style{}.Background(tcell.ColorDarkGray).Foreground(tcell.ColorGray))

		stop_btn.SetStyle(tcell.Style{}.Background(tcell.ColorRed).Foreground(tcell.ColorWhite))
		exec_btn.SetStyle(tcell.Style{}.Background(tcell.ColorDarkGreen).Foreground(tcell.ColorWhite))

		order_buttons := tview.NewGrid().SetColumns(20, 0, 20, 0, 20).
			AddItem(back_btn, 0, 0, 1, 1, 0, 0, false).
			AddItem(stop_btn, 0, 2, 1, 1, 0, 0, false).
			AddItem(exec_btn, 0, 4, 1, 1, 0, 0, false)

		cart_section = tview.NewGrid().SetRows(1, 0, 10, 1).
			AddItem(order_top, 0, 0, 1, 1, 0, 0, false).
			AddItem(checkout_table, 1, 0, 1, 1, 0, 0, false).
			AddItem(total_tbl, 2, 0, 1, 1, 2, 0, false).
			AddItem(order_buttons, 3, 0, 1, 1, 2, 0, false)

		entryPage.AddItem(cart_section, 1, 2, 1, 1, 0, 130, false)
	}

	pd_cred := func() {
		if order_form != nil {
			entryPage.RemoveItem(order_form)
		}

		sel2 := []string{"-- SELECT --"}
		credit_shop_items := api.GetCreditsItems()
		for _, v := range credit_shop_items {
			sel2 = append(sel2, *v.Name)
		}

		order_form = tview.NewForm().
			AddDropDown("Bundle Type", sel2, 0, func(option string, optionIndex int) {
				// find selected shop item
				credOrderData.ItemType = option

			}).
			AddInputField("Amount", "", 20, onlyNumbers, func(text string) {
				n, err := strconv.Atoi(text)
				if err == nil {
					credOrderData.Amount = n
				}
			}).
			AddButton("Cancel", func() {
				entryPage.RemoveItem(order_form).AddItem(order_config_basic, 1, 1, 1, 1, 0, 100, false)
				app.SetFocus(main_menu_list)
			}).
			AddButton("Order directly", func() {
				go func() {
					app.Draw()
					res, err := orderCredits(credit_shop_items, order_form)
					if err != nil {
						genericModal(fmt.Sprintf("Error: %s", err.Error()))
						return
					}
					app.Draw()
					if err != nil {
						genericModal(fmt.Sprintf("Error: %s", err.Error()))
						return
					}
					browserModal(res)
					app.Draw()
				}()
			})
		order_form.SetBorder(true).SetTitle("Order configuration").SetTitleAlign(tview.AlignCenter)
		entryPage.RemoveItem(order_config_basic).AddItem(order_form, 1, 1, 1, 1, 0, 100, false)
		app.SetFocus(order_form)
	}

	main_menu_list = tview.NewList().
		AddItem("Buy Basic Preplanning", "Browse basic preplanning assets", 'b', basic_sel).
		AddItem("Buy Exclusive Preplanning", "Browse heist-exclusive preplanning assets", 'e', exclusive_sel).
		AddItem("C-Stacks Marketplace", "Buy C-Stacks directly from the source", 's', gold_sel).
		AddItem("Add Credits", "Buy PayDay Credits from Nebula", 'c', pd_cred).
		AddItem("Quit", "Press to exit", 'q', func() {
			app.Stop()
		})

	entryPage = tview.NewGrid().
		SetRows(2, 0, 1).
		SetColumns(45, 45, 0).
		SetBorders(true).
		AddItem(newPrimitive(fmt.Sprintf("PayShop3 - Your Personal Black Market | %v", B_VER)), 2, 0, 1, 3, 0, 0, false)

	jumpToEntry := api.LD.DisplayName != ""

	entryPage.
		AddItem(main_menu_list, 1, 0, 1, 1, 0, 130, true).
		AddItem(order_config_basic, 1, 1, 1, 1, 0, 130, false).
		AddItem(cart_section, 1, 2, 1, 1, 0, 130, false)

	pages.AddPage("login", loginScreen, true, !jumpToEntry)
	pages.AddPage("entry", entryPage, true, jumpToEntry)

	if jumpToEntry {
		updateHeaderUI()
	}

	headerTimedUpdate()
	updateCartUI()
	if err := app.SetRoot(pages, true).SetFocus(pages).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
