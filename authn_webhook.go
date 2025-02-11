package main

import (
	"encoding/json"
	"io"
)

type authnWebhookRequest struct {
	RoomID        string       `json:"roomId"`
	ClientID      string       `json:"clientId"`
	ConnectionID  string       `json:"connectionId"`
	SignalingKey  *string      `json:"signalingKey,omitempty"`
	AuthnMetadata *interface{} `json:"authnMetadata,omitempty"`
	AyameClient   *string      `json:"ayameClient,omitempty"`
	Libwebrtc     *string      `json:"libwebrtc,omitempty"`
	Environment   *string      `json:"environment,omitempty"`
}

type authnWebhookResponse struct {
	Allowed       *bool        `json:"allowed"`
	IceServers    *[]iceServer `json:"iceServers"`
	Reason        *string      `json:"reason"`
	AuthzMetadata *interface{} `json:"authzMetadata"`
}

func (c *connection) authnWebhook() (*authnWebhookResponse, error) {
	if config.AuthnWebhookURL == "" {
		var allowed = true
		authnWebhookResponse := &authnWebhookResponse{Allowed: &allowed}
		return authnWebhookResponse, nil
	}

	req := &authnWebhookRequest{
		RoomID:        c.roomID,
		ClientID:      c.clientID,
		ConnectionID:  c.ID,
		SignalingKey:  c.signalingKey,
		AuthnMetadata: c.authnMetadata,
		AyameClient:   c.ayameClient,
		Libwebrtc:     c.libwebrtc,
		Environment:   c.environment,
	}

	resp, err := c.postRequest(config.AuthnWebhookURL, req)
	if err != nil {
		c.errLog().Err(err).Caller().Msg("AuthnWebhookError")
		return nil, errAuthnWebhook
	}
	// http://ikawaha.hateblo.jp/entry/2015/06/07/074155
	defer resp.Body.Close()

	c.webhookLog("authnReq", req)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.errLog().Bytes("body", body).Err(err).Caller().Msg("AuthnWebhookResponseError")
		return nil, err
	}

	// ログ出力用
	httpResponse := &httpResponse{
		Status: resp.Status,
		Proto:  resp.Proto,
		Header: resp.Header,
		Body:   string(body),
	}

	// 200 以外で返ってきたときはエラーとする
	if resp.StatusCode != 200 {
		c.errLog().Interface("resp", httpResponse).Caller().Msg("AuthnWebhookUnexpectedStatusCode")
		return nil, errAuthnWebhookUnexpectedStatusCode
	}

	c.webhookLog("authnResp", httpResponse)

	authnWebhookResponse := authnWebhookResponse{}
	if err := json.Unmarshal(body, &authnWebhookResponse); err != nil {
		c.errLog().Err(err).Caller().Msg("AuthnWebhookResponseError")
		return nil, errAuthnWebhookResponse
	}

	return &authnWebhookResponse, nil
}
