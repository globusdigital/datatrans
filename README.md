# datatrans

This Go package implements the new JSON based Datatrans API #golang #paymentprovider

## Documentation

https://api-reference.datatrans.ch

https://docs.datatrans.ch/docs

## Usage

```go

	c, err := datatrans.MakeClient(
	    datatrans.OptionHTTPRequestFn((&http.Client{
			Timeout: 30*time.Second,
		}).Do),
	    datatrans.OptionMerchant{
			InternalID: "",
			Server:     "https://api.sandbox.datatrans.com",
			MerchantID: "32234323242",
			Password:   "dbce0e6cfc012e475c843c1bbb0ca439a048fe8e",
		},
	    // add more merchants if you like
	    datatrans.OptionMerchant{
			InternalID: "B",
			Server:     "https://api.sandbox.datatrans.com",
			MerchantID: "78967896789",
			Password:   "e249002bc8e0c36dd89c393bfc7f7aa369c5842f",
		},
	)
	// uses the merchant B
	bc := c.WithMerchant("B")
	bc.Status("324234234")
	
	// uses default merchant
	c.Status("65784567")
```

### My request needs additional fields which aren't in the struct!

How can I extend the JSON data posted to datatrans?

```go
	ri := &datatrans.RequestInitialize{
		Currency:   "CHF",
		RefNo:      "234234",
		AutoSettle: true,
		Amount:     10023,
		Language:   "DE",
		CustomFields: map[string]interface{}{
			"TWI": map[string]interface{}{
				"alias": "ZGZhc2RmYXNkZmFzZGZhc2Q=",
			},
		},
	}
	data, err := datatrans.MarshalJSON(ri)
	// handle error
	// using TWI for Twint specific parameters
	const wantJSON = `{"TWI":{"alias":"ZGZhc2RmYXNkZmFzZGZhc2Q="},"amount":10023,"autoSettle":true,"currency":"CHF","language":"DE","refno":"234234"}`
	if string(data) != wantJSON {
		t.Errorf("\nWant: %s\nHave: %s", wantJSON, data)
	}

```

# License

Mozilla Public License Version 2.0
