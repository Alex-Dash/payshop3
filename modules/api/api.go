/*
PayShop3 - An Interactive Order-based System for PayDay3
Source: https://github.com/Alex-Dash/payshop3
Copyright (C) 2023  AlexDash
*/
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

const (
	nebula_base_url string = "https://nebula.starbreeze.com"
	basic_auth      string = "Basic MGIzYmZkZjVhMjVmNDUyZmJkMzNhMzYxMzNhMmRlYWI6"
	cid             string = "d682bcf949cb4744b3cd4295bbdd9fef"
)

var stealth_headers []header = []header{
	{Key: "User-Agent", Value: "PAYDAY3/++UE4+Release-4.27-CL-0 Windows/10.0.19045.1.256.64bit"},
	{Key: "Namespace", Value: "pd3"},
	{Key: "Game-Client-Version", Value: "1.0.0.0"},
	{Key: "Accelbyte-Sdk-Version", Value: "21.0.3"},
	{Key: "Accelbyte-Oss-Version", Value: "0.8.11"},
}

type LoginData struct {
	Login            string    `json:"nebula_user_login,omitempty"`
	Password         string    `json:"nebula_user_password,omitempty"`
	Token            string    `json:"access_token,omitempty"`
	RefreshToken     string    `json:"refresh_token,omitempty"`
	UserId           string    `json:"user_id,omitempty"`
	TokenType        string    `json:"token_type,omitempty"`
	TokenTTL         int       `json:"expires_in,omitempty"`
	RefreshTokenTTL  int       `json:"refresh_expires_in,omitempty"`
	DisplayName      string    `json:"display_name,omitempty"`
	AuthTrustId      string    `json:"auth_trust_id,omitempty"`
	TokenUpdatedAt   time.Time `json:"token_updated_at_pshop3,omitempty"`
	RefreshUpdatedAt time.Time `json:"refresh_updated_at_pshop3,omitempty"`
	AutoLogin        bool      `json:"autologin_pshop3,omitempty"`
}

type header struct {
	Key   string
	Value string
}

type ShopData struct {
	Data *[]ShopItemData `json:"data,omitempty"`
}

type ShopItemData struct {
	Title           *string           `json:"title,omitempty"`
	ItemId          *string           `json:"itemId,omitempty"`
	BaseAppId       *string           `json:"baseAppId,omitempty"`
	Sku             *string           `json:"sku,omitempty"`
	PrettyName      *string           `json:"pretty_name_pshop3,omitempty"`
	PrettyHeistName *string           `json:"pretty_heist_name_pshop3,omitempty"`
	Namespace       *string           `json:"namespace,omitempty"`
	Name            *string           `json:"name,omitempty"`
	EntitlementType *string           `json:"entitlementType,omitempty"`
	UseCount        *int              `json:"useCount,omitempty"`
	Stackable       *bool             `json:"stackable,omitempty"`
	CategoryPath    *string           `json:"categoryPath,omitempty"`
	Status          *string           `json:"status,omitempty"`
	Listable        *bool             `json:"listable,omitempty"`
	Purchasable     *bool             `json:"purchasable,omitempty"`
	ItemType        *string           `json:"itemType,omitempty"`
	RegionData      *[]ItemRegionData `json:"regionData,omitempty"`
	RegionDataItem  *ItemRegionData   `json:"regionDataItem,omitempty"`
	Images          *[]ItemImageData  `json:"images,omitempty"`
	ItemIds         *[]string         `json:"itemIds,omitempty"`
	ItemQty         *interface{}      `json:"itemQty,omitempty"`
	BoundItemIds    *[]string         `json:"boundItemIds,omitempty"`
	Tags            *[]string         `json:"tags,omitempty"`
	Features        *[]string         `json:"features,omitempty"`
	MaxCountPerUser *int              `json:"maxCountPerUser,omitempty"`
	MaxCount        *int              `json:"maxCount,omitempty"`
	Region          *string           `json:"region,omitempty"`
	Language        *string           `json:"language,omitempty"`
	CreatedAt       *string           `json:"createdAt,omitempty"`
	UpdatedAt       *string           `json:"updatedAt,omitempty"`
}

type ItemRegionData struct {
	Price              *int       `json:"price,omitempty"`
	DiscountPercentage *int       `json:"discountPercentage,omitempty"`
	DiscountAmount     *int       `json:"discountAmount,omitempty"`
	DiscountedPrice    *int       `json:"discountedPrice,omitempty"`
	CurrencyCode       *string    `json:"currencyCode,omitempty"`
	CurrencyType       *string    `json:"currencyType,omitempty"`
	CurrencyNamespace  *string    `json:"currencyNamespace,omitempty"`
	PurchaseAt         *time.Time `json:"purchaseAt,omitempty"`
	DiscountPurchaseAt *time.Time `json:"discountPurchaseAt,omitempty"`
}

type ItemImageData struct {
	As            *string `json:"as,omitempty"`
	Caption       *string `json:"caption,omitempty"`
	Height        *int    `json:"height,omitempty"`
	Width         *int    `json:"width,omitempty"`
	ImageUrl      *string `json:"imageUrl,omitempty"`
	SmallImageUrl *string `json:"smallImageUrl,omitempty"`
}

type OrderRespData struct {
	OrderNo              *string            `json:"orderNo,omitempty"`
	PaymentOrderNo       *string            `json:"paymentOrderNo,omitempty"`
	Namespace            *string            `json:"namespace,omitempty"`
	UserId               *string            `json:"userId,omitempty"`
	ItemId               *string            `json:"itemId,omitempty"`
	Sandbox              *bool              `json:"sandbox,omitempty"`
	Quantity             *int               `json:"quantity,omitempty"`
	Price                *int               `json:"price,omitempty"`
	Tax                  *int               `json:"tax,omitempty"`
	Vat                  *int               `json:"vat,omitempty"`
	SalesTax             *int               `json:"salesTax,omitempty"`
	PaymentProviderFee   *int               `json:"paymentProviderFee,omitempty"`
	PaymentMethodFee     *int               `json:"paymentMethodFee,omitempty"`
	Currency             *OrderCurrencyData `json:"currency,omitempty"`
	PaymentStationUrl    *string            `json:"paymentStationUrl,omitempty"`
	ItemSnapshot         *ShopItemData      `json:"itemSnapshot,omitempty"`
	Region               *string            `json:"region,omitempty"`
	Language             *string            `json:"language,omitempty"`
	Status               *string            `json:"status,omitempty"`
	CreatedTime          *time.Time         `json:"createdTime,omitempty"`
	ExpireTime           *time.Time         `json:"expireTime,omitempty"`
	PaymentRemainSeconds *int               `json:"paymentRemainSeconds,omitempty"`
	TotalTax             *int               `json:"totalTax,omitempty"`
	TotalPrice           *int               `json:"totalPrice,omitempty"`
	SubtotalPrice        *int               `json:"subtotalPrice,omitempty"`
	CreatedAt            *time.Time         `json:"createdAt,omitempty"`
	UpdatedAt            *time.Time         `json:"updatedAt,omitempty"`
}

type OrderCurrencyData struct {
	CurrencyCode   *string `json:"currencyCode,omitempty"`
	CurrencySymbol *string `json:"currencySymbol,omitempty"`
	CurrencyType   *string `json:"currencyType,omitempty"`
	Namespace      *string `json:"namespace,omitempty"`
	Decimals       *int    `json:"decimals,omitempty"`
}

type OrderInitData struct {
	ItemId          string `json:"itemId,omitempty"`
	Quantity        int    `json:"quantity,omitempty"`
	Price           int    `json:"price,omitempty"`
	DiscountedPrice int    `json:"discountedPrice,omitempty"`
	CurrencyCode    string `json:"currencyCode,omitempty"`
	Region          string `json:"region,omitempty"`
	Language        string `json:"language,omitempty"`
	ReturnUrl       string `json:"returnUrl,omitempty"`
	PrettyName      string `json:"pretty_name_pshop3,omitempty"`
	PrettyHeistName string `json:"pretty_heist_name_pshop3,omitempty"`
}

type OrderErrorData struct {
	ErrorCode    *int    `json:"errorCode,omitempty"`
	ErrorMessage *string `json:"errorMessage,omitempty"`
}

type AssetGroupData struct {
	GroupName  string
	PrettyName string
	Sku        string
	Bank       []ShopItemData
}

type WalletLinkedData struct {
	Id             *string `json:"id,omitempty"`
	Namespace      *string `json:"namespace,omitempty"`
	UserId         *string `json:"userId,omitempty"`
	CurrencyCode   *string `json:"currencyCode,omitempty"`
	CurrencySymbol *string `json:"currencySymbol,omitempty"`
	Balance        *int    `json:"balance,omitempty"`
	BalanceOrigin  *string `json:"balanceOrigin,omitempty"`
	// timeLimitedBalances - unknown array, skipped
	CreatedAt               *time.Time `json:"createdAt,omitempty"`
	UpdatedAt               *time.Time `json:"updatedAt,omitempty"`
	TotalPermanentBalance   *int       `json:"totalPermanentBalance,omitempty"`
	TotalTimeLimitedBalance *int       `json:"totalTimeLimitedBalance,omitempty"`
	Status                  *string    `json:"status,omitempty"`
}

type WalletData struct {
	Namespace      *string             `json:"namespace,omitempty"`
	UserId         *string             `json:"userId,omitempty"`
	CurrencyCode   *string             `json:"currencyCode,omitempty"`
	CurrencySymbol *string             `json:"currencySymbol,omitempty"`
	Balance        *int                `json:"balance,omitempty"`
	WalletInfos    *[]WalletLinkedData `json:"walletInfos,omitempty"`
	WalletStatus   *string             `json:"walletStatus,omitempty"`
	Status         *string             `json:"status,omitempty"`
	Id             *string             `json:"id,omitempty"`
}

type BasicOrderData struct {
	ItemTypeID int
	ItemType   string
	BuyTypeID  int
	BuyType    string
	Amount     int
}

type ExclusiveOrderData struct {
	HeistTypeID int
	HeistType   string
	ItemTypeID  int
	ItemType    string
	ItemTypeSKU string
	BuyTypeID   int
	BuyType     string
	Amount      int
}

type TokenClaims struct {
	Bans           *interface{}     `json:"bans,omitempty"`
	ClientId       *string          `json:"clinet_id,omitempty"`
	Country        *string          `json:"country,omitempty"`
	ExpiresAt      *int             `json:"exp,omitempty"`
	IssuedAt       *int             `json:"iat,omitempty"`
	IsComply       *bool            `json:"is_comply,omitempty"`
	Issuer         *string          `json:"iss,omitempty"`
	Flags          *int             `json:"jflgs,omitempty"`
	Namespace      *string          `json:"namespace,omitempty"`
	NamespaceRoles *[]NamespaceRole `json:"namespace_roles,omitempty"`
	Permissions    *interface{}     `json:"permissions,omitempty"`
	Roles          *[]string        `json:"roles,omitempty"`
	Scope          *string          `json:"scope,omitempty"`
	Sub            *string          `json:"sub,omitempty"`
	jwt.StandardClaims
}

type NamespaceRole struct {
	Namespace *string `json:"namespace,omitempty"`
	RoleId    *string `json:"role_id,omitempty"`
}

type GoldOrderData struct {
	BuyTypeID int
	BuyType   string
	Amount    int
}

var LD LoginData = LoginData{}
var Shop ShopData = ShopData{}
var Wallets []WalletData = []WalletData{}
var logout bool = true

// Initialize login details
func Init(login string, password string, save bool) error {
	if login == "" || password == "" {
		return errors.New("login or password cannot be empty")
	}
	LD = LoginData{}
	logout = false

	authResp, status, err := apicall("/iam/v3/oauth/token", "POST", []header{
		{Key: "Authorization", Value: basic_auth},
		{Key: "Content-Type", Value: "application/x-www-form-urlencoded;charset=UTF-8"},
	}, fmt.Sprintf("grant_type=password&client_id=%s&username=%s&password=%s&extend_exp=true", cid, login, password))

	if err != nil || status != 200 {
		logout = true
		return errors.New("login or password is incorrect")
	}

	err = json.Unmarshal(authResp, &LD)
	if err != nil {
		logout = true
		return errors.New("unexpected server response on attempted auth")
	}
	LD.TokenUpdatedAt = time.Now()
	LD.RefreshUpdatedAt = time.Now()
	LD.AutoLogin = save
	if save {
		LD.Password = password
		LD.Login = login
		savedata, _ := json.Marshal(LD)
		os.WriteFile("payshop3_logindata.json", savedata, 0644)
	} else {
		os.Remove("payshop3_logindata.json")
	}

	Shop, err = GetShop()
	if err != nil {
		logout = true
		return err
	}

	err = UpdateWallets()
	if err != nil {
		logout = true
		return err
	}

	// auto refresh tokens
	time.AfterFunc(time.Minute, func() { UpdateTokenInfo(save, true, false) })

	return nil
}

func GetShop() (ShopData, error) {
	var sd ShopData
	shopRaw, status, err := apicall("/platform/public/namespaces/pd3/items/byCriteria?limit=2147483647&includeSubCategoryItem=true", "GET", []header{}, "")

	if err != nil || status != 200 {
		return sd, errors.New("failed to query the shop")
	}

	err = json.Unmarshal(shopRaw, &sd)
	if err != nil {
		return sd, errors.New("failed to parse shop response")
	}

	return sd, nil
}

func UpdateShop() error {
	var err error
	Shop, err = GetShop()
	return err
}

func apicall(path string, method string, headers []header, body string) ([]byte, int, error) {
	req, err := http.NewRequest(method, nebula_base_url+path, strings.NewReader(body))
	if err != nil {
		return []byte{}, 0, err
	}

	// Main Headers
	for _, v := range headers {
		req.Header.Set(v.Key, v.Value)
	}

	// Stealth
	for _, v := range stealth_headers {
		req.Header.Set(v.Key, v.Value)
	}

	// Postlogin auth headers
	if LD.Token != "" {
		req.Header.Set("Authorization", LD.TokenType+" "+LD.Token)
		req.Header.Set("Cookie", fmt.Sprintf("access_token=%s; refresh_token=%s", LD.Token, LD.RefreshToken))
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte{}, 0, err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return []byte{}, 0, err
	}

	return resBody, res.StatusCode, nil
}

func LookupItemByIdLocal(id string) (ShopItemData, error) {
	for _, v := range *Shop.Data {
		if *v.ItemId == id {
			return v, nil
		}
	}
	return ShopItemData{}, errors.New("item with given id was not found locally")
}

func BuyItem(id string, quantity int) (OrderRespData, error) {
	item, err := LookupItemByIdLocal(id)
	if err != nil {
		return OrderRespData{}, err
	}
	rd := *item.RegionData

	body, _ := json.Marshal(OrderInitData{
		ItemId:          *item.ItemId,
		Quantity:        quantity,
		Price:           *rd[0].Price * quantity,
		DiscountedPrice: *rd[0].DiscountedPrice * quantity,
		CurrencyCode:    *rd[0].CurrencyCode,
		Region:          *item.Region,
		Language:        *item.Language,
		ReturnUrl:       "http://127.0.0.1",
	})

	orderRaw, status, err := apicall(fmt.Sprintf("/platform/public/namespaces/pd3/users/%s/orders", LD.UserId), "POST", []header{}, string(body))
	if err != nil || status != 201 {
		return OrderRespData{}, err
	}

	if status == 400 {
		// parse error
		e := OrderErrorData{}
		err = json.Unmarshal(orderRaw, &e)
		if err != nil {
			return OrderRespData{}, errors.New("failed to parse order error data")
		}
		if e.ErrorMessage != nil {
			// return payment order error
			return OrderRespData{}, errors.New(*e.ErrorMessage)
		}
		return OrderRespData{}, errors.New("failed to place an order")
	}

	order := OrderRespData{}
	err = json.Unmarshal(orderRaw, &order)
	if err != nil {
		return OrderRespData{}, errors.New("failed to read order response data")
	}
	return order, nil
}

func GetExclusivePreplanningAssets() []ShopItemData {
	arr := []ShopItemData{}
	for _, v := range *Shop.Data {
		if *v.CategoryPath == "/PreplanningAssets" {
			arr = append(arr, v)
		}
	}
	return arr
}

func GetAssetBank() []AssetGroupData {
	agd := []AssetGroupData{}
	for _, v := range *Shop.Data {
		if *v.CategoryPath != "/PreplanningAssets" {
			continue
		}
		sku := strings.Split(*v.Sku, "_")[2]
		f := false
		for i, s := range agd {
			if s.Sku == sku {
				agd[i].Bank = append(agd[i].Bank, v)
				f = true
				break
			}
		}
		if !f {
			t := []ShopItemData{v}
			agd = append(agd, AssetGroupData{Sku: sku, Bank: t})
		}
	}
	return agd
}

func UpdateWallets() error {
	arr := []string{"CASH", "GOLD", "CRED"}
	Wallets = []WalletData{}
	for _, w := range arr {
		var wd WalletData
		walletRaw, status, err := apicall(fmt.Sprintf("/platform/public/namespaces/pd3/users/%s/wallets/%s", LD.UserId, w), "GET", []header{}, "")

		if err != nil || status != 200 {
			return errors.New("failed to update wallets")
		}

		err = json.Unmarshal(walletRaw, &wd)
		if err != nil {
			return errors.New("failed to parse wallet response")
		}
		Wallets = append(Wallets, wd)
	}
	return nil
}

func GetCachedWalletByCode(c string) (WalletData, error) {
	for _, v := range Wallets {
		if *v.CurrencyCode == c {
			return v, nil
		}
	}
	return WalletData{}, fmt.Errorf("wallet %s could not be found", c)
}

func GetExclusiveAssetGroupBySku(sku string) AssetGroupData {

	for _, v := range GetAssetBank() {
		if v.Sku == sku {
			return v
		}
	}

	var agd AssetGroupData
	agd.Bank = []ShopItemData{}
	return agd
}

func UpdateTokenInfo(save bool, recursive bool, force bool) error {

	if logout {
		return nil
	}
	ttl_t := GetTimeLeftJWT(LD.Token) > 120
	ttl_rt := GetTimeLeftJWT(LD.RefreshToken) > 120

	if ttl_t && !recursive && !force {
		// tokens are valid, no need to update
		return nil
	}
	rt := LD.RefreshToken
	autologin := LD.AutoLogin
	lg := LD.Login
	pwd := LD.Password

	if !ttl_t && ttl_rt && !force {
		LD = LoginData{}

		authResp, status, err := apicall("/iam/v3/oauth/token", "POST", []header{
			{Key: "Authorization", Value: basic_auth},
			{Key: "Content-Type", Value: "application/x-www-form-urlencoded;charset=UTF-8"},
		}, fmt.Sprintf("grant_type=refresh_token&refresh_token=%s", rt))

		if err != nil || status != 200 {
			// try credentials
			force = true
		} else {
			// all good
			err = json.Unmarshal(authResp, &LD)
			if err != nil {
				return errors.New("unexpected server response on attempted token refresh")
			}
		}
	}

	if (!ttl_t && !ttl_rt) || force {
		// update via credentials
		LD = LoginData{}
		authResp, status, err := apicall("/iam/v3/oauth/token", "POST", []header{
			{Key: "Authorization", Value: basic_auth},
			{Key: "Content-Type", Value: "application/x-www-form-urlencoded;charset=UTF-8"},
		}, fmt.Sprintf("grant_type=password&client_id=%s&username=%s&password=%s&extend_exp=true", cid, lg, pwd))

		if err != nil || status != 200 {
			return errors.New("login or password is incorrect")
		}

		err = json.Unmarshal(authResp, &LD)
		if err != nil {
			return errors.New("unexpected server response on attempted auth")
		}
	}

	LD.TokenUpdatedAt = time.Now()
	LD.RefreshUpdatedAt = time.Now()
	LD.AutoLogin = autologin

	if save {
		// save new login info to file
		LD.Password = pwd
		LD.Login = lg
		savedata, _ := json.Marshal(LD)
		os.WriteFile("payshop3_logindata.json", savedata, 0644)
	}

	if !recursive {
		return nil
	}

	time.AfterFunc(time.Minute, func() { UpdateTokenInfo(save, true, false) })
	return nil
}

func GetTimeLeftJWT(tokenString string) int {
	cl := TokenClaims{}
	jwt.ParseWithClaims(tokenString, &cl, func(token *jwt.Token) (interface{}, error) {
		return []byte(""), nil
	})
	return *cl.ExpiresAt - int(time.Now().Unix())
}

func GetItemBySKU(sku string) (ShopItemData, error) {
	for _, v := range *Shop.Data {
		if *v.Sku == sku {
			return v, nil
		}
	}
	return ShopItemData{}, errors.New("item was not found in the shop")
}

func safeguard(itemid string) bool {
	for _, v := range *Shop.Data {
		if *v.ItemId == itemid {
			return *v.Purchasable && *v.Listable
		}
	}
	return false
}

func ExecOrder(item OrderInitData) (OrderRespData, error) {
	if !safeguard(item.ItemId) {
		return OrderRespData{}, errors.New("item was not found or not publicly avalible for purchase")
	}

	// Create a clear object so the server wouldn't get confused
	body, err := json.Marshal(OrderInitData{
		ItemId:          item.ItemId,
		Quantity:        item.Quantity,
		Price:           item.Price,
		DiscountedPrice: item.DiscountedPrice,
		CurrencyCode:    item.CurrencyCode,
		Region:          item.Region,
		Language:        item.Language,
		ReturnUrl:       item.ReturnUrl,
	})
	if err != nil {
		return OrderRespData{}, errors.New("failed to create order object")
	}
	orderResp, status, err := apicall(fmt.Sprintf("/platform/public/namespaces/pd3/users/%s/orders", LD.UserId), "POST", []header{
		{Key: "Content-Type", Value: "application/json"},
		{Key: "Accept", Value: "application/json"},
	}, string(body))
	if err != nil || status != 201 {
		if err == nil {
			var er OrderErrorData
			err_j := json.Unmarshal(orderResp, &er)
			if err_j != nil {
				fmt.Println(string(orderResp))
				return OrderRespData{}, err_j
			}
			return OrderRespData{}, errors.New(*er.ErrorMessage)
		}
		return OrderRespData{}, fmt.Errorf("failed to execute order for itemId: %v", item.ItemId)
	}

	var resp OrderRespData
	err_jr := json.Unmarshal(orderResp, &resp)
	if err_jr != nil {
		fmt.Println(string(orderResp))
		return OrderRespData{}, err_jr
	}

	return resp, nil
}

func Logout() {
	LD = LoginData{}
	Shop = ShopData{}
	Wallets = []WalletData{}
	logout = true
	os.Remove("payshop3_logindata.json")
}
