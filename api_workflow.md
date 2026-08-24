Here is the complete, step-by-step workflow for every endpoint in your updated **handle-first, passwordless Web3 backend architecture**.

Each section details what the endpoint does and how data flows through the application layers: **Middleware $\rightarrow$ Controller $\rightarrow$ Service $\rightarrow$ Repository $\rightarrow$ Database**.

---

### 1. Health Check (`GET /health`)

* **What it does:** Acts as a lightweight liveness probe to check if your backend server is online, running, and reachable.
* **How it works:**
1. The request hits the public `mux` router.
2. The inline handler responds instantly with HTTP status `200 OK` and body `"OK"`. It bypasses the database completely for speed.



---

### 2. Login / Sign-In (`POST /api/v1/auth/login`)

* **What it does:** Authenticates a returning user who types their `.nid` handle and signs a message with their wallet.
* **How it works:**
1. **Payload Receipt:** Client sends JSON containing `handle`, `address`, `signature`, `message`, and `chain`.
2. **Controller (`AuthController`):** Decodes the JSON body into a `LoginRequest` DTO.
3. **Service (`AuthService`):** Cleans and validates the input parameters.
4. **Repository (`AuthRepository`):**
* Queries the `handles` table to check if the handle exists.
* Verifies if the incoming wallet `address` matches the owner of that handle.
* If everything matches, retrieves the `user_id`.


5. **Token Generation:** Uses a helper utility (`helpers.GenerateToken`) to sign an HMAC session token containing the `user_id`.
6. **Response:** Returns the bearer token, `user_id`, and `handle` to the client.



---

### 3. Public Handle Resolution (`GET /api/v1/resolve`)

* **What it does:** Allows anyone on the web to look up a blockchain address mapped to a specific `.nid` handle and chain (e.g., finding the EVM address for `alice.nid`).
* **How it works:**
1. **Query Params:** Extracts `handle` and `chain` from the URL query string (`?handle=alice&chain=evm`).
2. **Controller (`ResolutionController`):** Ensures both query parameters are provided.
3. **Repository (`ResolutionRepository`):** Performs an SQL `JOIN` between `handles` and `wallets` where `handles.name = $1`, `wallets.chain = $2`, and status is active.
4. **Response:** Returns the mapped blockchain address or a `404 Not Found` if it doesn't exist.



---

### 4. Claim Handle / Homepage Registration (`POST /api/v1/handles/claim`) — Public

* **What it does:** The core homepage onboarding endpoint. A brand-new visitor types a handle, connects their wallet, signs a message, and instantly creates their account.
* **How it works:**
1. **Payload Receipt:** Client sends JSON with `name`, `address`, `chain`, `signature`, and `message`.
2. **Controller (`HandleController`):** Decodes the JSON payload.
3. **Service (`HandleService`):** Sanitizes the handle name (trimming spaces and converting to lowercase) and verifies required fields.
4. **Repository (`HandleRepository`):** Runs an atomic database transaction:
* Checks if the handle name is already taken.
* If the wallet address is brand new, creates a record in `users` and links it in `wallets`.
* Inserts the new handle into the `handles` table with `is_primary = true`.


5. **Response:** Returns the newly created handle metadata (`id`, `user_id`, `name`, `status`, `is_primary`, `created_at`).



---

### 5. Link Wallet (`POST /api/v1/wallets/link`) — Protected

* **What it does:** Allows an already authenticated user to link an additional blockchain address (e.g., adding a Solana wallet to an existing EVM-based `.nid` account).
* **How it works:**
1. **Auth Middleware:** Intercepts the request, validates the `Bearer Token` in the header, and injects the `userID` into the request context.
2. **Controller (`WalletController`):** Extracts `userID` from context and decodes the JSON payload (`chain`, `network`, `address`).
3. **Repository (`WalletRepository`):** Inserts the new wallet record into the `wallets` table linked to the user's `userID` with a `'verified'` status.
4. **Response:** Returns success and the linked wallet details.



---

### 6. Revoke Session (`POST /api/v1/sessions/revoke`) — Protected

* **What it does:** Allows a user to invalidate an active OAuth session or token login instance.
* **How it works:**
1. **Auth Middleware:** Validates identity via the bearer token and injects `userID`.
2. **Controller (`SessionController`):** Reads the target session ID from query parameters (`?id=...`).
3. **Repository (`SessionRepository`):** Executes an `UPDATE` query on the `oauth_sessions` table, expiring the session where `id = $1` and `user_id = $2` (ensuring users can only revoke their own sessions).
4. **Response:** Returns a confirmation message.



---

### 7. Get User Profile (`GET /api/v1/user/profile`) — Protected

* **What it does:** Aggregates a comprehensive dashboard view for the logged-in user, showing their account info, all active handles, and all linked wallets across chains.
* **How it works:**
1. **Auth Middleware:** Verifies the bearer token and supplies the `userID` through the request context.
2. **Controller (`UserController`):** Passes the `userID` to the service layer.
3. **Repository (`UserRepository`):** Executes targeted queries to fetch user account creation details, all associated handle names from `handles`, and all linked wallets from `wallets`.
4. **Service (`UserService`):** Formats and combines these results into a unified structure.
5. **Response:** Returns the clean JSON profile aggregate to the frontend dashboard.
