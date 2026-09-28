const API = (window.__API_URL__ || "http://localhost:8080/api").replace(/\/$/, "");
const $ = (id) => document.getElementById(id);
function message(text, type = "") { $("message").textContent = text; $("message").className = "msg" + (type ? ` ${type}` : ""); }

$("loginForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const response = await fetch(API + "/login", { method: "POST", headers: { "Content-Type": "application/json" }, credentials: "include", body: JSON.stringify({ email: $("email").value, password: $("password").value }) });
  const data = await response.json();
  if (!response.ok) { message(data.error || "Login failed.", "error"); return; }
  const profileResponse = await fetch(API + "/profile", { credentials: "include" });
  const profile = await profileResponse.json();
  $("accountName").textContent = profile.name || "Admin";
  $("loginForm").classList.add("hidden");
  $("adminPanel").classList.remove("hidden");
  document.querySelector(".auth-shell").classList.add("admin-authenticated");
  loadProducts();
});

async function loadProducts() {
  const response = await fetch(API + "/products", { credentials: "include" });
  const data = await response.json();
  if (!response.ok) { message(data.error || "Could not load products.", "error"); return; }
  $("productList").innerHTML = data.length ? data.map((product) => `<article class="product-item"><div><h3>${escapeHTML(product.name)}</h3><p>${escapeHTML(product.description || "No description")}</p><strong>$${Number(product.price).toFixed(2)}</strong><span class="quantity">Qty: ${product.quantity}</span></div><div class="product-item-actions"><button class="secondary-btn small-btn" type="button" data-edit="${product.id}">Edit</button><button class="danger-btn" type="button" data-delete="${product.id}">Delete</button></div></article>`).join("") : '<p class="empty-products">No products yet.</p>';
  document.querySelectorAll("[data-edit]").forEach((button) => button.addEventListener("click", () => editProduct(data.find((product) => product.id === button.dataset.edit))));
  document.querySelectorAll("[data-delete]").forEach((button) => button.addEventListener("click", () => deleteProduct(button.dataset.delete)));
}

async function loadOrders() {
  const response = await fetch(API + "/orders", { credentials: "include" });
  const data = await response.json();
  if (!response.ok) { message(data.error || "Could not load orders.", "error"); return; }
  $("orderList").innerHTML = data.length ? data.map((order) => `<article class="order-card"><div><span class="order-status">${escapeHTML(order.status)}</span><h3>Order #${escapeHTML(String(order.id).slice(-6))}</h3><p>${order.items.length} item(s) · ${new Date(order.createdAt).toLocaleDateString()}</p></div><strong>$${Number(order.total).toFixed(2)}</strong></article>`).join("") : '<p class="empty-products">No orders yet.</p>';
}

$("productForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const id = $("productId").value;
  const response = await fetch(API + (id ? `/products/${id}` : "/products"), { method: id ? "PUT" : "POST", headers: { "Content-Type": "application/json" }, credentials: "include", body: JSON.stringify({ name: $("productName").value, description: $("productDescription").value, price: Number($("productPrice").value), quantity: Number($("productQuantity").value) }) });
  const data = await response.json();
  if (!response.ok) { message(data.error || "Could not save product.", "error"); return; }
  resetForm();
  message(id ? "Product updated." : "Product added.", "success");
  loadProducts();
});

function editProduct(product) { showProductForm(); $("productId").value = product.id; $("productName").value = product.name; $("productDescription").value = product.description || ""; $("productPrice").value = product.price; $("productQuantity").value = product.quantity; $("productSubmit").textContent = "Update product"; $("productCancel").classList.remove("hidden"); }
function showProductForm() { $("productForm").classList.remove("hidden"); }
function resetForm() { $("productForm").reset(); $("productId").value = ""; $("productSubmit").textContent = "Add product"; $("productCancel").classList.add("hidden"); $("productForm").classList.add("hidden"); }
async function deleteProduct(id) { if (!confirm("Delete this product?")) return; const response = await fetch(API + `/products/${id}`, { method: "DELETE", credentials: "include" }); if (!response.ok) { message("Could not delete product.", "error"); return; } message("Product deleted.", "success"); loadProducts(); }
$("productCancel").addEventListener("click", resetForm);
$("openProductForm").addEventListener("click", showProductForm);
$("refreshProducts").addEventListener("click", loadProducts);
$("refreshOrders").addEventListener("click", loadOrders);
document.querySelectorAll("[data-view]").forEach((button) => button.addEventListener("click", () => {
  const ordersView = button.dataset.view === "orders";
  document.querySelectorAll("[data-view]").forEach((item) => item.classList.toggle("active", item === button));
  $("productsView").classList.toggle("hidden", ordersView);
  $("ordersView").classList.toggle("hidden", !ordersView);
  if (ordersView) loadOrders();
}));
$("logoutButton").addEventListener("click", async () => { await fetch(API + "/logout", { method: "POST", credentials: "include" }); window.location.reload(); });
function escapeHTML(value) { return String(value).replace(/[&<>'"]/g, (character) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "'": "&#39;", '"': "&quot;" }[character])); }
