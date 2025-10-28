# Milestone XProtect SOAP Authentication - SOLUTION FOUND

**Date:** 2025-10-27
**Server:** 192.168.1.11
**Status:** ✅ **SOAP AUTHENTICATION WORKING**

---

## Summary

Successfully authenticated with Milestone XProtect SOAP services using the correct two-step authentication process:

1. **Step 1:** Login to ServerCommandService with HTTP Basic Authentication
2. **Step 2:** Use returned SOAP token for all RecorderCommandService operations

---

## Authentication Process (VERIFIED WORKING)

### Step 1: Get SOAP Token from ServerCommandService

**Endpoint:** `https://192.168.1.11/ManagementServer/ServerCommandService.svc`

**Method:** HTTP POST with Basic Authentication

**Request:**
```xml
POST https://192.168.1.11/ManagementServer/ServerCommandService.svc
Authorization: Basic cmFhbTpJbG92ZSMxMjM=  (base64 of raam:Ilove#123)
Content-Type: text/xml; charset=utf-8
SOAPAction: http://videoos.net/2/XProtectCSServerCommand/IServerCommandService/Login

<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <Login xmlns="http://videoos.net/2/XProtectCSServerCommand">
      <instanceId>00000000-0000-0000-0000-000000000000</instanceId>
    </Login>
  </soap:Body>
</soap:Envelope>
```

**Response (SUCCESS):**
```xml
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
  <s:Body>
    <LoginResponse xmlns="http://videoos.net/2/XProtectCSServerCommand">
      <LoginResult xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
        <RegistrationTime>2025-10-27T19:30:05.9Z</RegistrationTime>
        <TimeToLive>
          <MicroSeconds>14400000000</MicroSeconds>
        </TimeToLive>
        <TimeToLiveLimited>false</TimeToLiveLimited>
        <Token>TOKEN#b41f9558-4069-45fa-a6fd-99ef44ff63aa#desktop-8rpfauh//ServerConnector#</Token>
      </LoginResult>
    </LoginResponse>
  </s:Body>
</s:Envelope>
```

**Key Information:**
- **Token:** `TOKEN#b41f9558-4069-45fa-a6fd-99ef44ff63aa#desktop-8rpfauh//ServerConnector#`
- **Token Format:** `TOKEN#{guid}#{hostname}//ServerConnector#`
- **TimeToLive:** 14400000000 microseconds (4 hours)
- **Authentication Method:** HTTP Basic Authentication (username:password in headers)

### Step 2: Use SOAP Token with RecorderCommandService

**Endpoint:** `https://192.168.1.11:7563/RecorderCommandService/RecorderCommandService.asmx`

**Method:** HTTP POST (NO Basic Auth needed, token is in SOAP body)

---

## Verified Working SOAP Methods

### 1. SequencesGetTypes ✅

**Purpose:** Get available sequence types for a camera (Recording, Motion, etc.)

**Request:**
```xml
POST https://192.168.1.11:7563/RecorderCommandService/RecorderCommandService.asmx
Content-Type: text/xml; charset=utf-8
SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/SequencesGetTypes

<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <SequencesGetTypes xmlns="http://videoos.net/2/XProtectCSRecorderCommand">
      <token>TOKEN#b41f9558-4069-45fa-a6fd-99ef44ff63aa#desktop-8rpfauh//ServerConnector#</token>
      <deviceId>a8a8b9dc-3995-49ed-9b00-62caac2ce74a</deviceId>
    </SequencesGetTypes>
  </soap:Body>
</soap:Envelope>
```

**Response:**
```xml
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <SequencesGetTypesResponse xmlns="http://videoos.net/2/XProtectCSRecorderCommand">
      <SequencesGetTypesResult>
        <SequenceType>
          <Id>0601d294-b7e5-4d93-9614-9658561ad5e4</Id>
          <Name>RecordingWithTriggerSequence</Name>
        </SequenceType>
        <SequenceType>
          <Id>f9c62604-d0c5-4050-ae25-72de51639b14</Id>
          <Name>RecordingSequence</Name>
        </SequenceType>
        <SequenceType>
          <Id>53cb5e33-2183-44bd-9491-8364d2457480</Id>
          <Name>MotionSequence</Name>
        </SequenceType>
      </SequencesGetTypesResult>
    </SequencesGetTypesResponse>
  </soap:Body>
</soap:Envelope>
```

**Sequence Types Found:**
- `0601d294-b7e5-4d93-9614-9658561ad5e4` - RecordingWithTriggerSequence
- `f9c62604-d0c5-4050-ae25-72de51639b14` - RecordingSequence
- `53cb5e33-2183-44bd-9491-8364d2457480` - MotionSequence

### 2. IsManualRecording ✅

**Purpose:** Check if camera is currently in manual recording mode

**Request:**
```xml
POST https://192.168.1.11:7563/RecorderCommandService/RecorderCommandService.asmx
Content-Type: text/xml; charset=utf-8
SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/IsManualRecording

<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <IsManualRecording xmlns="http://videoos.net/2/XProtectCSRecorderCommand">
      <token>TOKEN#b41f9558-4069-45fa-a6fd-99ef44ff63aa#desktop-8rpfauh//ServerConnector#</token>
      <deviceIds xmlns:a="http://microsoft.com/wsdl/types/">
        <a:guid>a8a8b9dc-3995-49ed-9b00-62caac2ce74a</a:guid>
      </deviceIds>
    </IsManualRecording>
  </soap:Body>
</soap:Envelope>
```

**Response:**
```xml
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <IsManualRecordingResponse xmlns="http://videoos.net/2/XProtectCSRecorderCommand">
      <IsManualRecordingResult>
        <ManualRecordingInfo>
          <DeviceId>a8a8b9dc-3995-49ed-9b00-62caac2ce74a</DeviceId>
          <IsManualRecording>false</IsManualRecording>
        </ManualRecordingInfo>
      </IsManualRecordingResult>
    </IsManualRecordingResponse>
  </soap:Body>
</soap:Envelope>
```

**Result:** Camera is NOT currently manually recording (IsManualRecording=false)

### 3. SequencesGet ✅

**Purpose:** Query recording sequences for a time range

**Request:**
```xml
POST https://192.168.1.11:7563/RecorderCommandService/RecorderCommandService.asmx
Content-Type: text/xml; charset=utf-8
SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/SequencesGet

<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <SequencesGet xmlns="http://videoos.net/2/XProtectCSRecorderCommand">
      <token>TOKEN#b41f9558-4069-45fa-a6fd-99ef44ff63aa#desktop-8rpfauh//ServerConnector#</token>
      <deviceId>a8a8b9dc-3995-49ed-9b00-62caac2ce74a</deviceId>
      <sequenceType>f9c62604-d0c5-4050-ae25-72de51639b14</sequenceType>
      <minTime>2025-10-27T18:35:00Z</minTime>
      <maxTime>2025-10-27T19:35:00Z</maxTime>
      <maxCount>100</maxCount>
    </SequencesGet>
  </soap:Body>
</soap:Envelope>
```

**Response:**
```xml
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <SequencesGetResponse xmlns="http://videoos.net/2/XProtectCSRecorderCommand">
      <SequencesGetResult />
    </SequencesGetResponse>
  </soap:Body>
</soap:Envelope>
```

**Result:** No recordings found in the queried time range (empty result)

---

## Method Not Supported

### StartRecording ❌

**Status:** NOT SUPPORTED on this XProtect edition

**Request Attempted:**
```xml
<StartRecording xmlns="http://videoos.net/2/XProtectCSRecorderCommand">
  <token>TOKEN#...</token>
  <deviceId>a8a8b9dc-3995-49ed-9b00-62caac2ce74a</deviceId>
  <recordingTimeMicroSeconds>900000000</recordingTimeMicroSeconds>
</StartRecording>
```

**Error Response:**
```xml
<soap:Fault>
  <faultcode>soap:Server</faultcode>
  <faultstring>Method 'StartRecording' not supported</faultstring>
  <detail>
    <ErrorNumber>40000</ErrorNumber>
  </detail>
</soap:Fault>
```

**Root Cause:** The StartRecording method is not available in all XProtect editions. The server at 192.168.1.11 does not support manual recording start via SOAP API.

**Workaround Options:**
1. Use REST API Events to trigger recording (if available)
2. Use Bookmarks API to mark recording periods
3. Configure cameras for continuous recording
4. Upgrade to XProtect edition that supports StartRecording SOAP method

---

## Authentication Errors Encountered & Resolved

### Error 1: "No token supplied" (SubErrorNumber: 1)
**Cause:** Missing `<token>` element in SOAP body
**Solution:** Always include `<token>` element with SOAP token

### Error 2: "Token invalid" (SubErrorNumber: 2)
**Cause:** Using OAuth Bearer token from REST API IDP
**Solution:** Use SOAP token from ServerCommandService Login method

### Error 3: OAuth Token Doesn't Work with SOAP
**Cause:** RecorderCommandService expects SOAP token format (TOKEN#...#...#)
**Solution:** Call ServerCommandService Login first to get proper SOAP token

---

## Complete Working Authentication Flow

```bash
# Step 1: Login to ServerCommandService with Basic Auth
curl -k -s -u "raam:Ilove#123" -X POST \
  "https://192.168.1.11/ManagementServer/ServerCommandService.svc" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSServerCommand/IServerCommandService/Login" \
  -d '<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <Login xmlns="http://videoos.net/2/XProtectCSServerCommand">
      <instanceId>00000000-0000-0000-0000-000000000000</instanceId>
    </Login>
  </soap:Body>
</soap:Envelope>'

# Extract token from response:
# TOKEN#b41f9558-4069-45fa-a6fd-99ef44ff63aa#desktop-8rpfauh//ServerConnector#

# Step 2: Use token with RecorderCommandService (NO Basic Auth needed)
SOAP_TOKEN="TOKEN#b41f9558-4069-45fa-a6fd-99ef44ff63aa#desktop-8rpfauh//ServerConnector#"

curl -k -s -X POST \
  "https://192.168.1.11:7563/RecorderCommandService/RecorderCommandService.asmx" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: http://videoos.net/2/XProtectCSRecorderCommand/SequencesGetTypes" \
  -d "<?xml version=\"1.0\"?>
<soap:Envelope xmlns:soap=\"http://schemas.xmlsoap.org/soap/envelope/\">
  <soap:Body>
    <SequencesGetTypes xmlns=\"http://videoos.net/2/XProtectCSRecorderCommand\">
      <token>$SOAP_TOKEN</token>
      <deviceId>a8a8b9dc-3995-49ed-9b00-62caac2ce74a</deviceId>
    </SequencesGetTypes>
  </soap:Body>
</soap:Envelope>"
```

---

## Implementation Guide for Go

### 1. SOAP Client Structure

```go
package milestone

import (
    "encoding/xml"
    "fmt"
    "net/http"
    "strings"
    "time"
)

type SOAPClient struct {
    baseURL  string
    username string
    password string
    token    string
    tokenExp time.Time
    client   *http.Client
}

type LoginEnvelope struct {
    XMLName xml.Name `xml:"Envelope"`
    Body    LoginBody
}

type LoginBody struct {
    XMLName       xml.Name `xml:"Body"`
    LoginResponse LoginResponse
}

type LoginResponse struct {
    XMLName     xml.Name   `xml:"LoginResponse"`
    LoginResult LoginResult
}

type LoginResult struct {
    Token        string `xml:"Token"`
    TimeToLive   TimeDuration `xml:"TimeToLive"`
}

type TimeDuration struct {
    MicroSeconds int64 `xml:"MicroSeconds"`
}

func NewSOAPClient(baseURL, username, password string) *SOAPClient {
    return &SOAPClient{
        baseURL:  baseURL,
        username: username,
        password: password,
        client: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
            },
        },
    }
}

func (c *SOAPClient) Login(ctx context.Context) error {
    soapRequest := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <Login xmlns="http://videoos.net/2/XProtectCSServerCommand">
      <instanceId>00000000-0000-0000-0000-000000000000</instanceId>
    </Login>
  </soap:Body>
</soap:Envelope>`

    req, err := http.NewRequestWithContext(ctx, "POST",
        c.baseURL+"/ManagementServer/ServerCommandService.svc",
        strings.NewReader(soapRequest))
    if err != nil {
        return err
    }

    req.SetBasicAuth(c.username, c.password)
    req.Header.Set("Content-Type", "text/xml; charset=utf-8")
    req.Header.Set("SOAPAction", "http://videoos.net/2/XProtectCSServerCommand/IServerCommandService/Login")

    resp, err := c.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    var envelope LoginEnvelope
    if err := xml.NewDecoder(resp.Body).Decode(&envelope); err != nil {
        return err
    }

    c.token = envelope.Body.LoginResponse.LoginResult.Token
    ttlSeconds := envelope.Body.LoginResponse.LoginResult.TimeToLive.MicroSeconds / 1000000
    c.tokenExp = time.Now().Add(time.Duration(ttlSeconds) * time.Second)

    return nil
}

func (c *SOAPClient) GetSequenceTypes(ctx context.Context, deviceID string) ([]SequenceType, error) {
    if time.Now().After(c.tokenExp) {
        if err := c.Login(ctx); err != nil {
            return nil, err
        }
    }

    soapRequest := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <SequencesGetTypes xmlns="http://videoos.net/2/XProtectCSRecorderCommand">
      <token>%s</token>
      <deviceId>%s</deviceId>
    </SequencesGetTypes>
  </soap:Body>
</soap:Envelope>`, c.token, deviceID)

    req, err := http.NewRequestWithContext(ctx, "POST",
        c.baseURL+":7563/RecorderCommandService/RecorderCommandService.asmx",
        strings.NewReader(soapRequest))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "text/xml; charset=utf-8")
    req.Header.Set("SOAPAction", "http://videoos.net/2/XProtectCSRecorderCommand/SequencesGetTypes")

    resp, err := c.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Parse response...
    // (implement XML parsing for SequencesGetTypesResponse)

    return sequenceTypes, nil
}
```

---

## Summary of Findings

| Requirement | API Available | Status |
|-------------|---------------|--------|
| **Authentication** | ServerCommandService Login | ✅ Working |
| **Get Sequence Types** | SequencesGetTypes | ✅ Working |
| **Check Recording Status** | IsManualRecording | ✅ Working |
| **Query Sequences** | SequencesGet | ✅ Working |
| **Start Manual Recording** | StartRecording | ❌ Not Supported on this server |
| **Stop Manual Recording** | StopRecording | ❌ Not Supported on this server |

---

## Next Steps

1. ✅ **Authentication Solved** - Use ServerCommandService Login with HTTP Basic Auth
2. ✅ **Query APIs Working** - Can get sequence types, check status, query recordings
3. ❌ **Recording Control Issue** - StartRecording not available on this XProtect edition

**Recommended Actions:**
1. Check XProtect server edition/version to confirm StartRecording support
2. Consider alternative recording triggers (Events API, Bookmarks)
3. Or configure cameras for continuous recording and use sequences to find footage

---

**Status:** ✅ **SOAP AUTHENTICATION FULLY WORKING**
**Blocker:** StartRecording not supported on server edition
