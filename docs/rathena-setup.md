# rAthena Setup

Goro currently targets the classic pre-renewal `2008-09-10aSakexe` flow.
Stock rAthena is close, but it needs a few compatibility patches for this
client profile.

## Required Patches

In `src/custom/defines_pre.hpp`, build rAthena as a pre-renewal Sakexe client:

```cpp
#define PACKETVER 20080910
#define PACKETVER_SAK_NUM 20080910
#define PRERE
```

Patch the character list layout for legacy Sakexe clients:

- `src/common/packets.hpp`: add a `PACKETVER_SAK_LEGACY_CHARINFO` path for
  pre-2009 Sakexe clients, where `CHARACTER_INFO` is 108 bytes.
- `src/char/char.cpp`: write HP/SP, slot and hair color using the legacy field
  sizes.
- `src/char/char_clif.cpp`: skip `HC_BLOCK_CHARACTER` for this legacy layout.

Patch packet handling:

- `src/map/clif_packetdb.hpp`: keep Renewal packet table entries guarded by
  Renewal packet macros, so the 2008 Sakexe profile does not inherit Renewal
  packet lengths.
- `src/map/packets_struct.hpp`: make `PACKETVER == 20080910` use the classic
  `ZC_USESKILL_ACK` cast packet layout.

## Required Config

In `conf/char_athena.conf`:

```conf
char_del_delay: 0
char_del_option: 1
pincode_enabled: no
char_move_enabled: no
allowed_job_flag: 1
```

`char_del_delay: 0` is important: delayed character deletion is for newer
clients, while the 2008 client deletes directly after the email check.

## Local Development Notes

- Point `conf/grf-files.txt` at the same 2008-era data used by Goro.
- Regenerate `map_cache.dat` if your GRF/data set does not match rAthena's
  default maps.
- For LAN testing, set `char_ip` and `map_ip` to the host IP that Goro should
  connect to.
- Higher EXP/drop rates, starter zeny, test NPCs, and GM accounts are optional
  convenience settings, not client compatibility requirements.
