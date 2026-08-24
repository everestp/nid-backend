Here is the complete, chronological story of how a user interacts with **NID.xyz** from the moment they arrive on your homepage, broken down into the two core workflows: **New User Onboarding** and **Returning User Sign-In**.

---

## Part 1: The New User Onboarding (First-Time Visitor)

This is what happens when someone visits your website for the very first time and wants to claim their permanent Web3 identity.

### Step-by-Step User Action & Backend Flow:

1. **The User Arrives:**
* The user opens `nid.xyz`. They see a clean homepage with an input box reading *"Claim your .nid name"*.
* They type in their desired name—for example, `everest`.


2. **Connecting the Wallet:**
* They click **"Get Started"** or **"Claim"**.
* Instead of a password or email form, their browser's Web3 wallet extension (like MetaMask for EVM or Phantom for Solana) automatically prompts them.


3. **Signing the Proof of Ownership:**
* The wallet pops up with a secure message to sign (e.g., *"Sign this message to claim everest.nid"*).
* The user clicks **Approve**. This cryptographic signature proves they actually control that blockchain address without exposing any passwords.


4. **The API Request:**
* The frontend bundles everything into a JSON payload and sends it to your backend:
```json
POST /api/v1/handles/claim
{
  "name": "everest",
  "address": "0x1234...5678",
  "chain": "evm",
  "signature": "0xfakesignature...",
  "message": "Claim everest.nid"
}

```




5. **What the Backend Does (`HandleController` $\rightarrow$ `Service` $\rightarrow$ `Repository`):**
* **Availability Check:** Queries the database (`handles` table) to ensure `everest` isn't already taken. If it is, it stops and returns an error.
* **User Creation (Atomic Transaction):**
* It checks if the wallet address already exists. Since it's a new user, it inserts a brand-new row into the `users` table and gets back a `user_id`.
* It inserts the wallet address into the `wallets` table, linked to that `user_id` with a `'verified'` status.
* It inserts `everest` into the `handles` table, marking it as `is_primary = true` and tying it to that same `user_id`.


* **Commit & Response:** Commits the transaction and returns the success response to the frontend.



**Result:** The user's account is live, their handle is locked in, and their wallet is permanently bound—all with zero passwords.

---

## Part 2: The Returning User Sign-In ("Sign in with NID")

This is what happens when a registered user comes back a week later to check their dashboard or log into an external website that uses NID.

### Step-by-Step User Action & Backend Flow:

1. **Entering the Handle:**
* The user clicks "Sign In" and types `everest` into the input field.


2. **Automatic Wallet Prompt:**
* Because the system knows `everest` is tied to a specific wallet address (or the frontend resolves it), the wallet extension pops up instantly requesting a login signature (e.g., *"Sign in to NID as everest"*).
* The user clicks **Approve**.


3. **The API Request:**
* The frontend sends the verification details to your backend login endpoint:
```json
POST /api/v1/auth/login
{
  "handle": "everest",
  "address": "0x1234...5678",
  "signature": "0xfakesignature...",
  "message": "Sign in to NID",
  "chain": "evm"
}

```




4. **What the Backend Does (`AuthController` $\rightarrow$ `Service` $\rightarrow$ `Repository`):**
* **Handle Verification:** Queries the `handles` table to find the owner of `everest`.
* **Security Check:** Verifies that the incoming wallet `address` actually belongs to the user who owns `everest`. If a malicious user tries to log in with a fake wallet under someone else's handle, the backend rejects it with an error ("this handle is already claimed by another wallet address").
* **Token Generation:** Once validated, your backend helper (`helpers.GenerateToken`) issues a secure session **Bearer Token** containing their `user_id`.


5. **Successful Entry:**
* The backend responds with:
```json
{
  "token": "eyJhbGciOi...",
  "user_id": "uuid-1234...",
  "handle": "everest"
}

```


* The frontend saves this token in local storage and redirects the user straight into their protected dashboard (`/api/v1/user/profile`).
