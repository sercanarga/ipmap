# ipmap ![release](https://badgen.net/github/release/sercanarga/ipmap) ![download](https://badgen.net/github/assets-dl/sercanarga/ipmap/Latest) ![license](https://badgen.net/github/license/sercanarga/ipmap)

ipmap is an open source, cross-platform and powerful network analysis tool.

## Installation
- Download the latest version from the [release page](https://github.com/sercanarga/ipmap/releases).
- Extract the downloaded zip file from the archive `unzip ipmap.zip`
- Set the file permission `chmod +x ipmap`
- Run ipmap with `./ipmap`

## Using ipmap
Run `./ipmap` to get all parameters

#### Parameters
```bash
-asn AS13335                         # get all IP blocks of ASN
-ip 103.21.244.0/22,103.22.200.0/22  # scans for entered IP blocks
-d example.com                       # search for domain in ASN/IP blocks
-t 200                               # timeout(ms) for requests (default:auto)
--c                                  # work until finish scanning
--export                             # auto export results (txt)
```

#### Usages
```bash
ipmap -asn AS13335 -t 300                                  # finding sites by scanning all the IP blocks in the ASN
ipmap -asn AS13335 -d example.com                          # finding real IP address of site by scanning all IP blocks in ASN
ipmap -ip 103.21.244.0/22,103.22.200.0/22 -t 300           # finding sites by scanning all the IP blocks
ipmap -ip 103.21.244.0/22,103.22.200.0/22 -d example.com   # finding real IP address of site by scanning given IP addresses
```

## Build
Clone the project

```bash
git clone https://github.com/sercanarga/ipmap.git
```

Go to directory and run go build
```bash
cd ipmap
go build .
```

Run the ipmap file in the directory
```bash
./ipmap
```

## Contributors
Thanks go to these wonderful people
<table>
  <tbody>
    <tr>
      <td align="center">
        <a href="https://github.com/ertugrulturan">
          <img src="https://avatars.githubusercontent.com/u/60829297?v=4" width="100px;" alt=""/>
          <br />
          <sub>
            <b>Ertuğrul TURAN</b>
          </sub>
        </a>
      </td>
      <td align="center">
        <a href="https://github.com/sametcodes">
          <img src="https://avatars.githubusercontent.com/u/9467273?v=4" width="100px;" alt=""/>
          <br />
          <sub>
            <b>Samet</b>
          </sub>
        </a>
      </td
    </tr>
  </tbody>
</table>
