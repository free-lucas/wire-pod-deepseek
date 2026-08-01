package sdkapp

import (
	"bytes"
	"crypto/tls"
	"net/http"
)

var transCfg = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore SSL warnings
}

func postSettingsUpdate(robot Robot, payload []byte) {
	url := "https://" + robot.Target + "/v1/update_settings"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+robot.GUID)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func setCustomEyeColor(robot Robot, hue string, sat string) {
	var updateJSON = []byte(`{"update_settings": true, "settings": {"custom_eye_color": {"enabled": true, "hue": ` + hue + `, "saturation": ` + sat + `} } }`)
	postSettingsUpdate(robot, updateJSON)
}

func setPresetEyeColor(robot Robot, value string) {
	var updateJSON = []byte(`{"update_settings": true, "settings": {"custom_eye_color": {"enabled": false}, "eye_color": ` + value + `} }`)
	postSettingsUpdate(robot, updateJSON)
}

func setSettingSDKstring(robot Robot, setting string, value string) {
	var updateJSON = []byte(`{"update_settings": true, "settings": {"` + setting + `": "` + value + `" } }`)
	postSettingsUpdate(robot, updateJSON)
}

func setSettingSDKintbool(robot Robot, setting string, value string) {
	var updateJSON = []byte(`{"update_settings": true, "settings": {"` + setting + `": ` + value + ` } }`)
	postSettingsUpdate(robot, updateJSON)
}

func setLocaleSetting(robot Robot, locale string) {
	setSettingSDKstring(robot, "locale", locale)
}

func setButtonWakewordSetting(robot Robot, value string) {
	setSettingSDKintbool(robot, "button_wakeword", value)
}
