const status = document.querySelector("#status");
const items = document.querySelector("#items");

async function loadItems() {
  try {
    const response = await fetch("/api/items");
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    const data = await response.json();
    items.replaceChildren(
      ...data.map((item) => {
        const element = document.createElement("li");
        element.textContent = `${item.id}: ${item.name}`;
        return element;
      }),
    );
    status.textContent = `${data.length} записей загружено из базы данных`;
  } catch (error) {
    status.textContent = `Ошибка загрузки: ${error.message}`;
  }
}

loadItems();
