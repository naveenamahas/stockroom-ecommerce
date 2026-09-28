const API = (window.__API_URL__ || "http://localhost:8080/api").replace(/\/$/, "");
const $ = (id) => document.getElementById(id);
const PRODUCT_IMAGES = [
  "https://images.unsplash.com/photo-1496181133206-80ce9b88a853?auto=format&fit=crop&w=800&q=80",
  "https://images.unsplash.com/photo-1511707171634-5f897ff02aa9?auto=format&fit=crop&w=800&q=80",
  "https://images.unsplash.com/photo-1523275335684-37898b6baf30?auto=format&fit=crop&w=800&q=80"
];
let shopProducts = [];
let checkoutProduct = null;

function message(text, type = "") {
  $("message").textContent = text;
  $("message").className = "msg" + (type ? ` ${type}` : "");
}

function toast(text, type = "") {
  const toastElement = $("toast");
  toastElement.textContent = text;
  toastElement.className = `toast visible ${type}`;
  window.clearTimeout(toast.timeout);
  toast.timeout = window.setTimeout(() => toastElement.classList.remove("visible"), 3200);
}

function setAuthPanel(panel) {
  $("loginForm").classList.toggle("hidden", panel !== "login");
  $("signupForm").classList.toggle("hidden", panel !== "signup");
  document.querySelectorAll(".tab").forEach((tab) => tab.classList.toggle("active", tab.dataset.panel === panel));
  message("");
}

document.querySelectorAll(".tab").forEach((tab) => tab.addEventListener("click", () => setAuthPanel(tab.dataset.panel)));

$("signupForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const response = await fetch(API + "/signup", { method: "POST", headers: { "Content-Type": "application/json" }, credentials: "include", body: JSON.stringify({ name: $("name").value, email: $("signupEmail").value, password: $("signupPassword").value }) });
  const data = await response.json();
  if (!response.ok) { message(data.error || "Signup failed.", "error"); return; }
  $("signupForm").reset();
  setAuthPanel("login");
  message("Account created. You can now login.", "success");
});

$("loginForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const response = await fetch(API + "/login", { method: "POST", headers: { "Content-Type": "application/json" }, credentials: "include", body: JSON.stringify({ email: $("email").value, password: $("password").value }) });
  const data = await response.json();
  if (!response.ok) { message(data.error || "Login failed.", "error"); return; }
  $("loginForm").classList.add("hidden");
  document.querySelector(".tabs").classList.add("hidden");
  $("shopPanel").classList.remove("hidden");
  document.querySelector(".auth-shell").classList.add("shop-authenticated");
  const profile = await fetch(API + "/profile", { credentials: "include" });
  const profileData = await profile.json();
  $("accountName").textContent = profileData.name || "Shopper";
  loadShopProducts();
  loadOrders();
});

async function loadShopProducts() {
  const response = await fetch(API + "/shop/products", { credentials: "include" });
  const data = await response.json();
  if (!response.ok) { message(data.error || "Could not load products.", "error"); return; }
  shopProducts = data;
  $("shopList").innerHTML = data.length ? data.map((product, index) => `<article class="shop-card"><div class="shop-card-image" style="background-image:url('${PRODUCT_IMAGES[index % PRODUCT_IMAGES.length]}')"><span>${escapeHTML(product.name)}</span></div><div class="shop-card-body"><div class="shop-card-top"><span class="shop-tag">STOCKROOM</span><span class="shop-stock ${product.quantity < 1 ? "sold-out" : ""}">${product.quantity < 1 ? "Out of stock" : `${product.quantity} available`}</span></div><h3>${escapeHTML(product.name)}</h3><p>${escapeHTML(product.description || "Quality product")}</p><div class="shop-card-bottom"><strong>$${Number(product.price).toFixed(2)}</strong><button class="primary-btn small-btn" type="button" ${product.quantity < 1 ? "disabled" : ""} data-id="${product.id}" data-quantity="${product.quantity}">Buy now</button></div></div></article>`).join("") : '<p class="empty-products">No products are available.</p>';
  document.querySelectorAll("#shopList button[data-id]").forEach((button) => button.addEventListener("click", () => openCheckout(button.dataset.id)));
}

function openCheckout(productId) {
  checkoutProduct = shopProducts.find((product) => product.id === productId);
  if (!checkoutProduct) return;
  $("checkoutProductName").textContent = checkoutProduct.name;
  $("checkoutProductDescription").textContent = checkoutProduct.description || "Quality product";
  $("checkoutProductPrice").textContent = `$${Number(checkoutProduct.price).toFixed(2)}`;
  $("checkoutQuantity").value = 1;
  $("checkoutQuantity").max = checkoutProduct.quantity;
  updateCheckoutTotal();
  $("checkout").classList.remove("hidden");
  $("checkout").scrollIntoView({ behavior: "smooth", block: "center" });
}

function updateCheckoutTotal() {
  if (!checkoutProduct) return;
  $("checkoutTotal").textContent = `$${(Number(checkoutProduct.price) * Number($("checkoutQuantity").value || 1)).toFixed(2)}`;
}

async function buyProduct(productId, quantity) {
  const response = await fetch(API + "/orders", { method: "POST", headers: { "Content-Type": "application/json" }, credentials: "include", body: JSON.stringify({ productId, quantity, paymentMethod: "dummy" }) });
  const data = await response.json();
  if (!response.ok) { toast(data.error || "Order could not be created.", "error"); loadShopProducts(); return; }
  $("confirmationText").textContent = `Order ${data.id.slice(-6).toUpperCase()} is confirmed. Your dummy payment was accepted.`;
  $("confirmationDetails").innerHTML = `<span>${escapeHTML(data.items[0].name)} × ${data.items[0].quantity}</span><strong>$${Number(data.total).toFixed(2)}</strong>`;
  $("checkout").classList.add("hidden");
  $("confirmation").classList.remove("hidden");
  toast("Payment successful. Your order is confirmed.", "success");
  loadShopProducts();
  loadOrders();
}

async function loadOrders() {
  const response = await fetch(API + "/orders", { credentials: "include" });
  const data = await response.json();
  if (!response.ok) return;
  $("ordersList").innerHTML = data.length ? data.map((order, index) => `<article class="order-card"><div class="order-image order-image-${index % 3}">${escapeHTML(order.items[0]?.name || "Order")}</div><div class="order-details"><span class="shop-tag">${escapeHTML(order.status.toUpperCase())}</span><h3>${escapeHTML(order.items.map((item) => `${item.name} x ${item.quantity}`).join(", "))}</h3><p>${new Date(order.createdAt).toLocaleString()}</p></div><strong>$${Number(order.total).toFixed(2)}</strong></article>`).join("") : '<p class="empty-products">No orders yet.</p>';
}

$("refreshOrders").addEventListener("click", loadOrders);
$("checkoutQuantity").addEventListener("input", updateCheckoutTotal);
$("closeCheckout").addEventListener("click", () => $("checkout").classList.add("hidden"));
$("cancelCheckout").addEventListener("click", () => $("checkout").classList.add("hidden"));
$("continueShopping").addEventListener("click", () => { $("confirmation").classList.add("hidden"); $("shopList").scrollIntoView({ behavior: "smooth", block: "start" }); });
$("payButton").addEventListener("click", () => {
  if (!checkoutProduct) return;
  const quantity = Number($("checkoutQuantity").value);
  if (!Number.isInteger(quantity) || quantity < 1 || quantity > checkoutProduct.quantity) { toast(`Choose a quantity from 1 to ${checkoutProduct.quantity}.`, "error"); return; }
  $("payButton").disabled = true;
  buyProduct(checkoutProduct.id, quantity).finally(() => { $("payButton").disabled = false; });
});
$("logoutButton").addEventListener("click", async () => { await fetch(API + "/logout", { method: "POST", credentials: "include" }); window.location.reload(); });
function escapeHTML(value) { return String(value).replace(/[&<>'"]/g, (character) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "'": "&#39;", '"': "&quot;" }[character])); }
