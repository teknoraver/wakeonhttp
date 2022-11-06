/*
 * woh - Wake-on-HTTP
 * HTTP endpoint to send Wake-on-lan packets
 * Copyright (C) 2022 Matteo Croce <technoboy85@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

func sendWol(hwaddr net.HardwareAddr, w http.ResponseWriter) {
	/* Since there is no ARP entry wich would allow us to reach the
	 * destination machine, send a broadcast packet.
	 * UDP port 9 was historically used for a test service named 'discard'
	 * which just ignores data, so it's safe to use.
	 */
	conn, err := net.Dial("udp", "255.255.255.255:9")
	if err != nil {
		http.Error(w, "Socket error", 500)
		return
	}
	defer conn.Close()

	/* The "magic packet" is composed by six 0xff and then the MAC
	 * of the machine to wake up repeated 16 times.
	 * Total size of the UDP payload is 102 octects (6+6*16)
	 */
	p := make([]byte, 102)
	for i := 0; i < 6; i++ {
		p[i] = 0xff
	}
	for i := 1; i < 17; i++ {
		copy(p[i*6:i*6+6], hwaddr)
	}
	_, err =  conn.Write(p)
	if err != nil {
		http.Error(w, "I/O error", 500)
		return
	}
}

func parseAddr(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	addr, present := params["addr"]
	if !present {
		http.Error(w, "missing 'addr' argument", 400)
		return
	}
	hwaddr, err := net.ParseMAC(addr[0])
	if err != nil {
		http.Error(w, fmt.Sprintf("'%s' is not a valid MAC address", addr[0]), 400)
		return
	}
	sendWol(hwaddr, w)
}

func main() {
	fmt.Println()
	http.HandleFunc("/wake", parseAddr)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
