---
title: Descripción general de la arquitectura
description: Arquitectura de alto nivel de Stellar Explorer.
---

Stellar Explorer es una aplicación Next.js 16 que lee datos de la red Stellar a través de dos APIs: Horizon (REST) y Soroban RPC (JSON-RPC).

![Architecture Overview](../../../../assets/diagrams/architecture-overview.svg)

## Componentes

### Frontend

El frontend es una aplicación React 19 construida con Next.js App Router. Usa TanStack Query para la obtención y caché de datos, shadcn/ui como biblioteca de componentes, y Tailwind CSS 4 para los estilos.

Todas las páginas se renderizan en el servidor en la primera carga y luego se hidratan para la navegación del lado del cliente.

### Red Stellar

La aplicación se conecta directamente a la infraestructura pública de Stellar:

- **Horizon API** — API REST para transacciones, cuentas, activos, ledgers y efectos
- **Soroban RPC** — JSON-RPC para datos de contratos inteligentes (código, almacenamiento, eventos)

No se requiere un backend personalizado para la funcionalidad básica del explorador.

### Servicios de backend (indexador)

Un servicio indexador opcional basado en Go procesa datos del ledger de Stellar en PostgreSQL (con TimescaleDB), Redis y Typesense para consultas avanzadas, búsqueda de texto completo y analíticas. Ver [Indexador](./indexer/) para más detalles.

Stellar Explorer también incluye una interfaz terminal en `apps/tui` y un servicio de datos dedicado en `services/tui-indexer`. Juntos permiten búsqueda de datos Stellar con teclado, monitoreo en vivo, búsqueda indexada, timelines y lecturas con origen visible para flujos terminales.

## Decisiones de diseño clave

- **Sin servidor API personalizado** — El frontend consulta Stellar directamente, reduciendo infraestructura
- **Agnóstico de red** — Toda la obtención de datos incluye el parámetro de red
- **Clientes SDK en caché** — Los clientes Horizon y RPC se crean una vez por red y se reutilizan
- **Caché de datos inmutables** — Ledgers, transacciones y código de contratos usan `staleTime: Infinity`
