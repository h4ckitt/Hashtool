## HASHTOOL

Pre-Requisite Task Written For Bidding Eligibility In HNG-i9 Internship


## Installation

### Requirements
 - go 1.17+
 
Build The Tool With The Following Command:
```shell
$ go build -o hashtool hashtool.go
```

Run The Resulting Binary As Follows:
```shell
$ ./hashtool <inputfile.csv>
```

## Things To Note
 
- Data Was Sanitized (e.g. Extra Spaces And Leading Trimmed) Before The SHASUM Was Computed.

The Following Assumptions Were Made As There Wasn't Any Clarity On It:

- The Collection ID Remains The Same For All NFTs Since All The NFTs Are Part Of The Same Collection Which Is `Zuri NFT Tickets for Free Lunch`
- The Minting Tool Is The Name Of Each Team That The NFT Belongs To Which Is Specified In The HNG-i9 CSV
- NFTs That Do Not Have A Name Or Other Attributes Are Not To Be Processed.


## Sample Data
The Following JSON Is A Sample Output For NFT Number 61 From `TEAM GRIT`
```json
{
  "format":"CHIP-0007",
  "name":"toy-soldier",
  "description":"a man stronger than an army",
  "minting_tool":"TEAM GRIT",
  "sensitive_content":false,
  "series_number":61,
  "series_total":420,
  "attributes":[
    {
      "trait_type":"gender",
      "value":"male"
    },
    {
      "trait_type":"hair",
      "value":" bald"
    },
    {
      "trait_type":"strengths",
      "value":" powerful"
    },
    {
      "trait_type":"weakness",
      "value":" sentimental"
    }
  ],
  "collection": {
    "name":"Zuri NFT Tickets for Free Lunch",
    "id":"b774f676-c1d5-422e-beed-00ef5510c64d",
    "Attributes":{
      "type":"description",
      "value":"Rewards for accomplishments during HNGi9."
    }
  }
}
```

It Has A SHA256SUM Of :
```text
9077aa62de924e9005f92e0b803062f27f61b971241961861f79edfca35f0df5
```