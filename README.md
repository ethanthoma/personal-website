<h3 align="center">
    <img 
        src="https://raw.githubusercontent.com/ethanthoma/personal-website/main/cmd/webserver/public/favicon/android-chrome-512x512.png" 
        width="100"
        alt="Logo"/>
    <br/>
    <a href="https://www.ethanthoma.com/">Personal Website</a>
</h3>

<p align="center">
    <img src="https://img.shields.io/github/last-commit/ethanthoma/personal-website/main?style=for-the-badge&labelColor=%231f1d2e&color=%23c4a7e7">
    <img src="https://img.shields.io/github/actions/workflow/status/ethanthoma/personal-website/docker.yml?style=for-the-badge&labelColor=%231f1d2e&color=%239ccfd8">
    <img src="https://img.shields.io/github/languages/count/ethanthoma/personal-website?style=for-the-badge&labelColor=%231f1d2e&color=%23ebbcba">
</p>


## GoTH Stack

| Tech  | Stack    |
|-------|----------|
| Go    | Backend  |
| Htmx  | Frontend |
| Turso | Database |

## Building + Running

The nix flake has three derivations:
- .#default: this produces the webserver binary
- .#container: docker image containing the webserver binary
- .#uploader: simple CLI to upload my markdown blogs

You can run it test the webserver locally with docker with `make run`. 
