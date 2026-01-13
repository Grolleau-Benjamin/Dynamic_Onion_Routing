# Onion Packet — Average Size Calculation

## Variables & Mean Values

Since the route length and the number of keys can vary per packet, we rely on mean values for this estimation.

- Route Dimensions (Averages):
  - $\bar{n}_j = 3$ (Average number of Jumps/Layers)
  - $\bar{n}_{nh} = 2$ (Average number of Next Hops per layer)
  - $S_{nh} = 19 \text{bytes}$ (Fixed NextHop Size: IPv6 + Type + Port)
- Crypto & Headers (Constants):
  - $S_{epk} = 32 \text{bytes}$ (Ephemeral Public Key)
  - $n_{wk}$= 3 (Fixed number of Wrapped Keys)
  - $S_{wk} = 76 \text{bytes}$ (Fixed Wrapped Key Size)
  - $S_{nonce} = 12 \text{bytes}$
  - $S_{meta} = 3 \text{bytes}$ (Flag + CipherTextLenXOR)
  - $S_{pl\_len} = 2 \text{bytes}$

## Average Overhead per Layer

We calculate the average "cost" added by a single relay node.

### Average Inner Metadata ($O_{inner}$)

Data added inside the *Ciphered Onion Layer* which is used as *CipheredPayload* in *Onion Layer*.

$$
O_{inner} = S_{meta} + S_{pl\_len} + (\bar{n}_{nh} * S_{nh})
$$

$$
O_{inner} = 3 + 2 + (2*19)
$$

$$
O_{inner} = 43 \text{bytes}
$$

### Outer Metadata ($O_{outer}$)

Data added *outside* the encryption (EPK, Wrapped Keys, Nonce). The *Onion Layer*‘s metadata.

$$
O_{outer} = S_{epk} + (n_{wk} * S_{wk}) + S_{meta} + S_{nonce}
$$

$$
O_{outer} = 32 + (3 * 76) + 3 + 12
$$

$$
O_{outer} = 275 \text{bytes}
$$

### Mean Layer Overhead ($O_{layer}$)

$$
O_{layer} = O_{outer} + O_{inner}
$$

$$
O_{layer} = 275 + 43
$$

$$
O_{layer} = 318 \text{bytes}
$$

## Total Packet Size Estimation

The estimated total size is the sum of the average overheads for all jumps plus the actual payload.

$$
\text{Est. Total Size} = (\bar{n}_j * O_{layer}) + \text{Payload Size}
$$

Numerical application

$$
\text{Est. Total Size} = (3 * 318) + \text{Payload Size}
$$

$$
\text{Est. Total Size} = 954 + \text{Payload Size}
$$
