# Sources

https://m-bus.com/documentation-wired/06-application-layer

# First Byte 0xE5

Confirmation of a previous sent frame.

# First Byte 0x10 - Short Frame

1. 0x10 (start)
2. CI-Field
3. A-Field
4. checksum
5. 0x16 (stop)

# First Byte 0x68 - Long Frame

1. 0x68 (start)
2. L-Field - length of the Data + 3 (C-Field, A-Field, CI-Field)
3. L-Field - length of the Data + 3 (C-Field, A-Field, CI-Field)
4. 0x68 (start)
5. C-Field
6. A-Field
7. CI-Field
8. DATA
9. checksum
10. 0x16 (stop)

# First Byte 0x68 - Control Frame

1. 0x68 (start)
2. L-Field = 0x03
3. L-Field = 0x03
4. 0x68 (start)
5. C-Field
6. A-Field
7. CI-Field
8. checksum
9. 0x16 (stop)

# L-Field

# C-Field

Control Field and Function field
Calling direction (master to slave) or responding direction (slave to master).

# A-Field

Address Field

# CI-Field

Control Information Field

The CI-Field codes the type and sequence of application data to be transmitted in this frame.

- slaves selection
- baudrate
- ...

# Communication Process

- Send/Confirm: SND/CON
- Request/Response: REQ/RSP

Wait between 11 bit times and (330 bit times + 50ms) before answering.

# Example

- The master switches the slave (in point to point connection) from now 2400 baud to 9600 baud.

```
Master to slave: 68 03 03 68 | 53 FE BD | 0E 16 with 2400 baud
Slave to master: E5 with 2400 baud
```

From that time on the slave communicates with the transmission speed 9600 baud.

- The master releases an enhanced application reset to all slaves. All telegrams of the user data type are requested.

```
Master to Slave: 68 04 04 68 | 53 FE 50 | 10 | B1 16
Slave to Master: E5
```

The slave with address 5 and identification number 12345678 responds with the following data (all values hex.):

```
68 13 13 68 header of RSP_UD telegram (L-Field = 13h = 19d)
08 05 73 C field = 08h (RSP_UD), address 5, CI field = 73h (fixed, LSByte first)
78 56 34 12 identification number = 12345678
0A transmission counter = 0Ah = 10d
00 status 00h: counters coded BCD, actual values, no errors
E9 7E Type&Unit: medium water, unit1 = 1l, unit2 = 1l (same, but historic)
01 00 00 00 counter 1 = 1l (actual value)
35 01 00 00 counter 2 = 135 l (historic value)
3C 16 checksum and stop sign
```

- Example for a RSP_UD with variable data structure answer (mode 1)

```
68 1F 1F 68 header of RSP_UD telegram (length 1Fh=31d bytes)
08 02 72 C field = 08 (RSP), address 2, CI field 72H (var.,LSByte first)
78 56 34 12 identification number = 12345678
24 40 01 07 manufacturer ID = 4024h (PAD in EN 61107), generation 1, water
55 00 00 00 TC = 55h = 85d, Status = 00h, Signature = 0000h
03 13 15 31 00 Data block 1: unit 0, storage No 0, no tariff, instantaneous volume,
12565 l (24 bit integer)
DA 02 3B 13 01 Data block 2: unit 0, storage No 5, no tariff, maximum volume flow,
113 l/h (4 digit BCD)
8B 60 04 37 18 02 Data block 3: unit 1, storage No 0, tariff 2, instantaneous energy,
218,37 kWh (6 digit BCD)
18 16 checksum and stopsign
```

- Set the slave to primary address 8 without changing anything else:

```
   68 06 06 68 | 53 FE 51 | 01 7A 08 | 25 16
```

- Set the complete identification of the slave (ID=01020304, Man=4024h (PAD), Gen=1, Med=4 (Heat):

```
68 0D 0D 68 | 53 FE 51 | 07 79 04 03 02 01 24 40 01 04 | 95 16 §
```

- Set identification number of the slave to "12345678" and the 8 digit BCD-Counter (unit 1 kWh) to 107 kWh.

```
68 0F 0F 68 | 53 FE 51| 0C 79 78 56 34 12 | 0C 06 07 01 00 00 | 55 16
```

- A slave with address 7 is to be configured to respond with the data records containing volume (VIF=13h: volume, unit
  1l) and flow temperature (VIF=5Ah: flow temp., unit 0.1 °C).

```
68 07 07 68 | 53 07 51 | 08 13 08 5A | 28 16
```

- A slave with address 1 is to be configured to respond with all storage numbers, all tariffs, and all VIF from unit 0.

```
68 06 06 68 | 53 01 51 | C8 3F 7E | 2A 16
```

- A slave with address 3 is to be configured to respond with all data for a complete readout of all available. After
  that the master can poll the slave to get the data.

```
68 04 04 68 | 53 03 51 | 7F | 26 16
```

- Set the 8 digit BCD-Counter (instantaneous, actual value, no tariff, unit 0) with VIF=06 (1kWh) of the slave with
  address 1 to 107 kWh.

```
68 0A 0A 68 | 53 01 51 | 0C 86 00 07 01 00 00 | 3F 16
```

- Same as in example 1) but add 10 kWh to the old data.

```
68 0A 0A 68 | 53 01 51 | 0C 86 01 10 00 00 00 | 48 16
```

- Add an entry with an 8 digit BCD-Counter (instantaneous, actual value, no tariff, unit 0, 1kWh) with the start value
  of 511 kWh to the data records of the slave with address 5.

```
68 0A 0A 68 | 53 05 51 | 0C 86 08 11 05 00 00 | 59 16
```

- Freeze actual flow temperature (0.1 °C: VIF = 5Ah) of the slave with address 1 into the storage number 1.

```
68 06 06 68 | 53 01 51 | 40 DA 0B | CA 16
```

- The fabrication number is a serial number allocated during manufacture. It is part of the variable data block (
  DIF = $0C and VIF = $78) and coded with 8 BCD packed digits (4 Byte).

```
68 15 15 68 header of RSP_UD telegram (length 1Fh=31d bytes)
08 02 72 C field = 08 (RSP), address 2, CI field 72H (var.,LSByte first)
78 56 34 12 identification number = 12345678
24 40 01 07 manufacturer ID = 4024h (PAD in EN 61107), generation 1, water
13 00 00 00 TC = 13h = 19d, Status = 00h, Signature = 0000h
0C 78 04 03 02 01 fabrication number = 01020304
9D 16 checksum and stopsign
```

