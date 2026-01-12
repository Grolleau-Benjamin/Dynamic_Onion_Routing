dor_proto = Proto("dor", "Dynamic Onion Routing")

-- Preferences
dor_proto.prefs.ports = Pref.string("TCP Ports", "62503,62504,62505,62506,62507,62508", "List of TCP ports to decode as DOR")

-- Protocol constants
local PACKET_SIZE = 4096
local MAX_WRAPPED_KEY = 3
local WRAPPED_KEY_SIZE = 76

-- Packet type constants
local TYPE_GET_IDENTITY_REQUEST = 0x00
local TYPE_GET_IDENTITY_RESPONSE = 0x01
local TYPE_ONION_PACKET = 0x10

-- Field definitions
local f_type = ProtoField.uint8("dor.type", "Packet Type", base.HEX, {
  [TYPE_GET_IDENTITY_REQUEST] = "GetIdentityRequest",
  [TYPE_GET_IDENTITY_RESPONSE] = "GetIdentityResponse",
  [TYPE_ONION_PACKET] = "OnionPacket",
})
local f_len = ProtoField.uint16("dor.length", "Payload Length", base.DEC)
local f_payload = ProtoField.bytes("dor.payload", "Payload")

-- Identity fields
local f_ruuid = ProtoField.bytes("dor.identity.ruuid", "Relay UUID", base.SPACE)
local f_pubkey = ProtoField.bytes("dor.identity.pubkey", "Public Key", base.SPACE)

-- Onion Layer fields
local f_onion_epk = ProtoField.bytes("dor.onion.epk", "Ephemeral Public Key", base.SPACE)
local f_onion_wrapped_keys = ProtoField.bytes("dor.onion.wrapped_keys", "Wrapped Keys", base.SPACE)
local f_onion_wk_nonce = ProtoField.bytes("dor.onion.wk.nonce", "Nonce", base.SPACE)
local f_onion_wk_cipher = ProtoField.bytes("dor.onion.wk.cipher", "Ciphertext", base.SPACE)
local f_onion_flags = ProtoField.uint8("dor.onion.flags", "Flags", base.HEX)
local f_onion_payload_nonce = ProtoField.bytes("dor.onion.payload_nonce", "Payload Nonce", base.SPACE)
local f_onion_ct_len_xor = ProtoField.uint16("dor.onion.ct_len_xor", "Ciphertext Length (XOR masked)", base.HEX)
local f_onion_ciphertext = ProtoField.bytes("dor.onion.ciphertext", "Ciphertext + Padding", base.SPACE)

dor_proto.fields = {
  f_type, f_len, f_payload,
  f_ruuid, f_pubkey,
  f_onion_epk, f_onion_wrapped_keys, f_onion_wk_nonce, f_onion_wk_cipher,
  f_onion_flags, f_onion_payload_nonce, f_onion_ct_len_xor,
  f_onion_ciphertext
}

-- Helpers
local function is_valid_msg_type(t)
  return t == TYPE_GET_IDENTITY_REQUEST or
         t == TYPE_GET_IDENTITY_RESPONSE or
         t == TYPE_ONION_PACKET
end

-- Dissect GetIdentityRequest (0x00)
local function dissect_msg_getidentityreq(tvb, pinfo, tree, plen)
  tree:set_text("GetIdentityRequest")

  if plen ~= 0 then
    tree:add_expert_info(PI_MALFORMED, PI_ERROR,
      string.format("GetIdentityRequest payload length must be 0, got %d", plen))
    return false
  end

  pinfo.cols.info = "DOR GetIdentityRequest"
  return true
end

-- Dissect GetIdentityResponse (0x01)
local function dissect_msg_getidentityres(tvb, pinfo, tree, plen)
  tree:set_text("GetIdentityResponse")

  if plen ~= 48 then
    tree:add_expert_info(PI_MALFORMED, PI_ERROR,
      string.format("GetIdentityResponse payload length must be 48, got %d", plen))
    return false
  end

  tree:add(f_ruuid, tvb(0, 16))
  tree:add(f_pubkey, tvb(16, 32))

  pinfo.cols.info = "DOR GetIdentityResponse"
  return true
end

-- Dissect OnionPacket (0x10)
local function dissect_msg_onionpacket(tvb, pinfo, tree, plen)
  tree:set_text(string.format("OnionPacket (%d bytes)", plen))

  if plen ~= PACKET_SIZE then
    tree:add_expert_info(PI_MALFORMED, PI_WARN,
      string.format("OnionPacket payload length should be %d, got %d", PACKET_SIZE, plen))
  end

  local offset = 0

  -- EPK (32 bytes)
  if offset + 32 <= plen then
    tree:add(f_onion_epk, tvb(offset, 32))
    offset = offset + 32
  else
    tree:add_expert_info(PI_MALFORMED, PI_ERROR, "Truncated: missing EPK")
    return false
  end

  -- Wrapped Keys (3 × 76 = 228 bytes)
  if offset + (MAX_WRAPPED_KEY * WRAPPED_KEY_SIZE) <= plen then
    local wk_tree = tree:add(f_onion_wrapped_keys, tvb(offset, MAX_WRAPPED_KEY * WRAPPED_KEY_SIZE))

    for i = 0, MAX_WRAPPED_KEY - 1 do
      local wk_offset = offset + (i * WRAPPED_KEY_SIZE)
      local wk_subtree = wk_tree:add(dor_proto, tvb(wk_offset, WRAPPED_KEY_SIZE),
        string.format("Wrapped Key [%d]", i))

      wk_subtree:add(f_onion_wk_nonce, tvb(wk_offset, 12))
      wk_subtree:add(f_onion_wk_cipher, tvb(wk_offset + 12, 64))
    end

    offset = offset + (MAX_WRAPPED_KEY * WRAPPED_KEY_SIZE)
  else
    tree:add_expert_info(PI_MALFORMED, PI_ERROR, "Truncated: missing Wrapped Keys")
    return false
  end

  -- Flags (1 byte)
  if offset + 1 <= plen then
    tree:add(f_onion_flags, tvb(offset, 1))
    offset = offset + 1
  else
    tree:add_expert_info(PI_MALFORMED, PI_ERROR, "Truncated: missing Flags")
    return false
  end

  -- Payload Nonce (12 bytes)
  if offset + 12 <= plen then
    tree:add(f_onion_payload_nonce, tvb(offset, 12))
    offset = offset + 12
  else
    tree:add_expert_info(PI_MALFORMED, PI_ERROR, "Truncated: missing Payload Nonce")
    return false
  end

  -- CipherText Length XOR (2 bytes)
  if offset + 2 <= plen then
    tree:add(f_onion_ct_len_xor, tvb(offset, 2))
    offset = offset + 2
  else
    tree:add_expert_info(PI_MALFORMED, PI_ERROR, "Truncated: missing CT Length XOR")
    return false
  end

  -- CipherText + Padding (rest of packet)
  local remaining = plen - offset
  if remaining > 0 then
    tree:add(f_onion_ciphertext, tvb(offset, remaining))
  end

  pinfo.cols.info = string.format("DOR OnionPacket (%d bytes)", plen)
  return true
end

local MIN_HDR = 3
local function dor_get_pdu_len(tvb, pinfo, offset)
  local tvb_len = tvb:len()

  if tvb_len < offset + MIN_HDR then
    return 0
  end

  local msg_type = tvb(offset, 1):uint()
  if not is_valid_msg_type(msg_type) then
    return 0
  end

  local plen = tvb(offset + 1, 2):uint()

  return MIN_HDR + plen
end

local function dor_dissect_one_pdu(tvb, pinfo, tree)
  local msg_type = tvb(0, 1):uint()
  local plen = tvb(1, 2):uint()
  local total_len = MIN_HDR + plen

  pinfo.cols.protocol = "DOR"

  local subtree = tree:add(dor_proto, tvb(0, total_len), "Dynamic Onion Routing Packet")
  subtree:add(f_type, tvb(0, 1))
  subtree:add(f_len, tvb(1, 2))

  if plen > 0 then
    local payload = tvb(3, plen):tvb()
    local paytree = subtree:add(f_payload, tvb(3, plen))

    if msg_type == TYPE_GET_IDENTITY_REQUEST then
      dissect_msg_getidentityreq(payload, pinfo, paytree, plen)
    elseif msg_type == TYPE_GET_IDENTITY_RESPONSE then
      dissect_msg_getidentityres(payload, pinfo, paytree, plen)
    elseif msg_type == TYPE_ONION_PACKET then
      dissect_msg_onionpacket(payload, pinfo, paytree, plen)
    else
      paytree:set_text(string.format("Unknown packet type 0x%02X", msg_type))
      pinfo.cols.info = string.format("DOR Unknown (0x%02X)", msg_type)
    end
  else
    if msg_type == TYPE_GET_IDENTITY_REQUEST then
      dissect_msg_getidentityreq(tvb(3, 0):tvb(), pinfo, subtree, 0)
    else
      pinfo.cols.info = string.format("DOR type 0x%02X (empty)", msg_type)
    end
  end

  return total_len
end

-- Main dissector function
function dor_proto.dissector(tvb, pinfo, tree)
  dissect_tcp_pdus(tvb, tree, MIN_HDR, dor_get_pdu_len, dor_dissect_one_pdu)
end

-- Register ports
local function register_ports()
  local tcp_port = DissectorTable.get("tcp.port")
  local port_string = dor_proto.prefs.ports

  for port_str in string.gmatch(port_string, "[^,]+") do
    port_str = port_str:match("^%s*(.-)%s*$")  -- Trim whitespace
    local port = tonumber(port_str)
    if port and port > 0 and port < 65536 then
      tcp_port:add(port, dor_proto)
    end
  end
end

register_ports()

-- Handle preference changes
function dor_proto.prefs_changed()
  local tcp_port = DissectorTable.get("tcp.port")
  tcp_port:remove_all(dor_proto)
  register_ports()
end

-- Plugin info
if set_plugin_info then
  set_plugin_info({
    version = "0.1.0",
    author = "DOR Protocol Team",
    description = "Dissector for Dynamic Onion Routing Protocol with detailed OnionPacket parsing",
    repository = "https://github.com/Grolleau-Benjamin/Dynamic_Onion_Routing"
  })
end
