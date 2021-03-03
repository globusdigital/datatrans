# datatrans

This Go package implements the new JSON based Datatrans API #golang #paymentprovider

## Documentation

https://api-reference.datatrans.ch

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

# License

Mozilla Public License Version 2.0
