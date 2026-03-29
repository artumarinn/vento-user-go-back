# Guía de Integración con Meta (WhatsApp, Messenger, Instagram)

Esta guía detalla los pasos y credenciales necesarios para conectar Vento.ai con la plataforma de Meta.

## 🛠️ Requisitos previos en Meta Developers

1.  **Crear una App**: Ir a [developers.facebook.com](https://developers.facebook.com), crear una App de tipo **"Business"**.
2.  **Agregar Productos**:
    *   **WhatsApp**: Para mensajería por WA.
    *   **Messenger**: Para Facebook e Instagram.
3.  **Configurar Webhooks**:
    *   URL: `https://tu-dominio.com/webhooks/meta`
    *   Verify Token: El valor que configures en `META_VERIFY_TOKEN` (ej: `vento-secret-handshake`).
    *   Fields (WhatsApp): `messages`.
    *   Fields (Messenger): `messages`, `messaging_postbacks`.

---

## 🔑 Credenciales Necesarias

Para que cada PyME funcione con Vento, necesitás guardar estos datos en la tabla `meta_configs` (a través del Dashboard en el futuro):

| Campo | Descripción | Dónde obtenerlo |
| :--- | :--- | :--- |
| `whatsapp_phone_number_id` | ID único del número de teléfono. | WhatsApp > Configuración de la API. |
| `whatsapp_business_id` | ID de la cuenta de WhatsApp Business. | WhatsApp > Configuración de la API. |
| `permanent_access_token` | Token de acceso de sistema (no expira). | Configuración del negocio > Usuarios del sistema. |
| `verify_token` | Token para el handshake inicial. | Lo inventás vos (debe coincidir con la App en Meta). |
| `app_secret` | Secreto de la App para validar firmas. | Configuración de la App > Información básica. |

---

## ⚙️ Configuración del Entorno (`.env`)

En el `ia-service`, asegurate de tener:
```env
META_VERIFY_TOKEN=tu-verify-token-global
META_APP_SECRET=tu-app-secret-de-la-app
```

---

## 🔄 Flujo de Mensajes

1.  **Meta** envía un POST a `/webhooks/meta`.
2.  **IA Service** valida la firma con `META_APP_SECRET`.
3.  **IA Service** busca el `phoneNumberID` en el **Core**.
4.  **Core** devuelve el Token y el `UserID` (para filtrar el RAG).
5.  **IA Service** consulta a Gemini con el contexto de los productos del usuario.
6.  **IA Service** responde al cliente usando el `permanent_access_token`.

---
© 2026 Vento.AI - Sistema de Mensajería Inteligente.
