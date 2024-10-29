package utils

const Banner = `
    _     _                  _  __                         ____       
   | |   (_)_ __  _   ___  _| |/ /___  ___ _ __   ___ _ / ___| ___  
   | |   | | '_ \| | | \ \/ / ' // _ \/ _ \ '_ \ / _ \ | |  _ / _ \ 
   | |___| | | | | |_| |>  <| . \  __/  __/ |_) |  __/ | |_| | (_) |
   |_____|_|_| |_|\__,_/_/\_\_|\_\___|\___| .__/ \___|  \____|\___/ 
                                          |_|                        
                                                                   
    [*] Linux Persistence Tool - Written in Go
    [*] Author: 10cks
    [*] Github: https://github.com/10cks/LinuxKeeperGo
    [*] Version: 1.0.0
`

func ShowBanner() {
	println(Banner)
}
