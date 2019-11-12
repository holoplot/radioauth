package main

import (
	"log"
	"reflect"

	"layeh.com/radius"
	"layeh.com/radius/rfc2759"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2868"
	"layeh.com/radius/rfc2869"
	"layeh.com/radius/rfc3079"
	"layeh.com/radius/vendors/microsoft"
)

func runRadiusServer() {
	handler := func(w radius.ResponseWriter, r *radius.Request) {
		username := rfc2865.UserName_GetString(r.Packet)
		challenge := microsoft.MSCHAPChallenge_Get(r.Packet)
		response := microsoft.MSCHAP2Response_Get(r.Packet)

		account, err := accountStore.Read(username)
		if err != nil {
			log.Printf("[radius] Cannot access account for %s: %v", username, err)
			w.Write(r.Response(radius.CodeAccessReject))
			return
		}

		if !authenticateToken(account) {
			log.Printf("[radius] Token authentication failed for %s", username)
			w.Write(r.Response(radius.CodeAccessReject))
			return
		}

		if len(challenge) == 16 && len(response) == 50 {
			// See rfc2548 - 2.3.2. MS-CHAP2-Response
			ident := response[0]
			peerChallenge := response[2:18]
			peerResponse := response[26:50]
			ntResponse, err := rfc2759.GenerateNTResponse(challenge, peerChallenge, username, account.Password)
			if err != nil {
				log.Printf("[radius] Cannot generate ntResponse for %s: %v", username, err)
				w.Write(r.Response(radius.CodeAccessReject))
				return
			}

			if reflect.DeepEqual(ntResponse, peerResponse) {
				responsePacket := r.Response(radius.CodeAccessAccept)

				recvKey, err := rfc3079.MakeKey(ntResponse, account.Password, false)
				if err != nil {
					log.Printf("[radius] Cannot make recvKey for %s: %v", username, err)
					w.Write(r.Response(radius.CodeAccessReject))
					return
				}

				sendKey, err := rfc3079.MakeKey(ntResponse, account.Password, true)
				if err != nil {
					log.Printf("[radius] Cannot make sendKey for %s: %v", username, err)
					w.Write(r.Response(radius.CodeAccessReject))
					return
				}

				authenticatorResponse, err := rfc2759.GenerateAuthenticatorResponse(challenge, peerChallenge, ntResponse, username, account.Password)
				if err != nil {
					log.Printf("[radius] Cannot generate authenticator response for %s: %v", username, err)
					w.Write(r.Response(radius.CodeAccessReject))
					return
				}

				success := make([]byte, 43)
				success[0] = ident
				copy(success[1:], authenticatorResponse)

				rfc2869.AcctInterimInterval_Add(responsePacket, rfc2869.AcctInterimInterval(3600))
				rfc2868.TunnelType_Add(responsePacket, 0, rfc2868.TunnelType_Value_L2TP)
				rfc2868.TunnelMediumType_Add(responsePacket, 0, rfc2868.TunnelMediumType_Value_IPv4)
				microsoft.MSCHAP2Success_Add(responsePacket, []byte(success))
				microsoft.MSMPPERecvKey_Add(responsePacket, recvKey)
				microsoft.MSMPPESendKey_Add(responsePacket, sendKey)
				microsoft.MSMPPEEncryptionPolicy_Add(responsePacket, microsoft.MSMPPEEncryptionPolicy_Value_EncryptionAllowed)
				microsoft.MSMPPEEncryptionTypes_Add(responsePacket, microsoft.MSMPPEEncryptionTypes_Value_RC440or128BitAllowed)

				log.Printf("[radius] Access granted to %s", username)
				w.Write(responsePacket)
				return
			}
		}

		log.Printf("[radius] Access denied for %s", username)
		w.Write(r.Response(radius.CodeAccessReject))
	}

	server := radius.PacketServer{
		Handler:      radius.HandlerFunc(handler),
		SecretSource: radius.StaticSecretSource([]byte(config.RadiusSecret)),
	}

	log.Printf("Starting Radius server on :1812")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
