---
title: Estructura del proyecto
description: Descripción general de la organización del código.
---

## Directorios de nivel superior

``` 
stellar-explorer/
├── apps/
│   ├── explorer-web/ # Frontend Next.js del explorador
│   ├── docs/         # Sitio de documentación Astro/Starlight
│   └── tui/          # Interfaz terminal de Stellar Explorer
├── services/
│   ├── indexer/      # Servicio Go estable de ingestión de datos
│   └── tui-indexer/  # Backend dedicado para flujos terminales
├── infra/
│   ├── docker/       # Configuración Docker
│   └── docker-compose.yml
└── .github/          # Workflows de CI
```

## Código fuente del frontend (`apps/explorer-web/src/`)

```
apps/explorer-web/src/
├── app/              # Páginas del App Router de Next.js
│   ├── [locale]/[network]/(explorer)/
│   └── api/          # Rutas API (fetcher TOML)
├── components/       # Componentes React
│   ├── ui/           # Componentes base de shadcn/ui
│   ├── layout/       # Encabezado, navegación, barras laterales
│   ├── cards/        # Tarjetas de transacción, operación, contrato
│   ├── charts/       # Visualizaciones con Recharts
│   ├── transactions/ # Componentes específicos de transacciones
│   ├── contracts/    # Componentes de contratos Soroban
│   ├── assets/       # Componentes de navegación de activos
│   ├── search/       # Interfaz de búsqueda
│   └── common/       # Componentes compartidos
├── lib/
│   ├── stellar/      # Clientes SDK y definiciones de consulta
│   ├── hooks/        # Hooks React personalizados
│   ├── providers/    # Proveedores de contexto
│   ├── constants/    # Constantes de la aplicación
│   ├── utils/        # Funciones de utilidad
│   └── types/        # Tipos TypeScript
└── i18n/             # Configuración de internacionalización
```

## Constantes clave

| Constante | Valor | Propósito |
|---|---|---|
| `STROOPS_PER_XLM` | `10.000.000` | Factor de conversión para cantidades XLM |
| `DEFAULT_PAGE_SIZE` | `20` | Elementos por lista paginada |
| `LIVE_LEDGER_POLL_INTERVAL` | `5.000 ms` | Intervalo de sondeo para datos en vivo |
| `STALE_TIME` | `10.000 ms` | Tiempo de caducidad predeterminado de TanStack Query |

## Stack de UI

- **Biblioteca de componentes:** shadcn/ui (estilo new-york)
- **Estilos:** Tailwind CSS 4
- **Gráficos:** Recharts
- **Alias de ruta:** `@/` apunta a `apps/explorer-web/src/`

## Producto terminal (`apps/tui/`)

`apps/tui` es la interfaz terminal de Stellar Explorer. Permite buscar y recorrer datos de Stellar con teclado, monitorear actividad en vivo, navegar entidades relacionadas, guardar contexto local y ver con claridad si cada resultado proviene de Stellar RPC o del backend dedicado del TUI.

### Tiers de tests del TUI

| Comando | Alcance | Requisitos |
|---|---|---|
| `bun run tui:test` | Suite completa del TUI (unit + integración) | Solo toolchain de Go |
| `bun run tui:test:unit` | Solo tests unitarios y de fixtures de render | Solo toolchain de Go |
| `bun run tui:test:integration` | Solo cadenas de confiabilidad (build tag `integration`) | Solo toolchain de Go |
| `bun run tui-indexer:test` | Handlers de read API y tests de store del indexer | Toolchain de Go; PostgreSQL para tests de integración del store |

Workflows de CI:

- `.github/workflows/tui-ci.yml` — build, lint, tests unitarios e integración de `apps/tui`
- `.github/workflows/tui-indexer-ci.yml` — build, lint, migraciones y tests de `services/tui-indexer`

## Backend de TUI (`services/tui-indexer/`)

`services/tui-indexer` prepara datos indexados de Stellar Explorer para flujos terminales. Expone APIs de lectura, búsqueda, timelines, registros relacionados y datos de actividad en vivo que enriquecen la experiencia del TUI más allá de consultas directas a Stellar RPC.
