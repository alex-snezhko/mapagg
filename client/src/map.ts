import L from "leaflet";
import type { OverlayBoundsResponse } from "./types";

export async function initMap() {
  const overlayBoundsRes = await fetch("http://localhost:8080/overlay-bounds");
  const overlayBounds = await overlayBoundsRes.json() as OverlayBoundsResponse;
  if (!overlayBounds.success) {
    alert("Failed to get overlay bounds");
    return;
  }

  const centerLat = (overlayBounds.data.bottomRight.lat + overlayBounds.data.topLeft.lat) / 2;
  const centerLong = (overlayBounds.data.bottomRight.long + overlayBounds.data.topLeft.long) / 2;
  const zoom = (overlayBounds.data.bottomRight.long - overlayBounds.data.topLeft.long) * 18;

  const map = L.map('map', { zoomControl: false }).setView([centerLat, centerLong], zoom);

  L.control.zoom({ position: 'topright' }).addTo(map);

  const CartoDB_Positron = L.tileLayer('https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png', {
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>',
    subdomains: 'abcd',
    maxZoom: 20
  });

  CartoDB_Positron.addTo(map);

  return map;
}
