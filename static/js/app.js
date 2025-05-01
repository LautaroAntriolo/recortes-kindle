// Función para configurar el selector de archivos
function configurarSelectorArchivos() {
    const selector = document.getElementById('selector-archivos');
    
    selector.addEventListener('change', async function() {
        const archivoSeleccionado = this.value;
        if (!archivoSeleccionado) {
            const tbody = document.getElementById('tabla-recortes');
            tbody.innerHTML = ''; // Limpiar tabla existente
        };
        
        try {
            // Enviar el nombre del archivo .txt (la API lo convertirá a .json)
            const response = await fetch(`/archivo/${encodeURIComponent(archivoSeleccionado)}`);
            
            if (!response.ok) {
                throw new Error(`Error ${response.status}: ${response.statusText}`);
            }
            
            // Primero obtener la respuesta como texto para poder inspeccionarla
            const textoRespuesta = await response.text();
            
            // Depuración - mostrar la respuesta completa para diagnóstico
            console.log("Respuesta completa del servidor:", textoRespuesta);
            
            // Intentar extraer solo la parte JSON de la respuesta
            let jsonLimpio;
            
            try {
                // Método 1: Buscar el primer carácter válido de JSON (normalmente '{' o '[')
                let inicioJSON = -1;
                const posibleInicio1 = textoRespuesta.indexOf('{');
                const posibleInicio2 = textoRespuesta.indexOf('[');
                
                if (posibleInicio1 >= 0 && (posibleInicio2 < 0 || posibleInicio1 < posibleInicio2)) {
                    inicioJSON = posibleInicio1;
                } else if (posibleInicio2 >= 0) {
                    inicioJSON = posibleInicio2;
                }
                
                if (inicioJSON > 0) {
                    console.log(`Encontrado texto no JSON al inicio. Eliminando los primeros ${inicioJSON} caracteres.`);
                    jsonLimpio = textoRespuesta.substring(inicioJSON);
                } else {
                    jsonLimpio = textoRespuesta;
                }
                
                // Método 2: Si el método 1 falla, intentar extraer usando expresiones regulares
                if (jsonLimpio.indexOf('{') < 0 && jsonLimpio.indexOf('[') < 0) {
                    const regexJSON = /[\[\{].*[\]\}]/s;
                    const coincidencias = textoRespuesta.match(regexJSON);
                    if (coincidencias && coincidencias[0]) {
                        console.log("Extrayendo JSON usando regex");
                        jsonLimpio = coincidencias[0];
                    }
                }
                
                // Método 3: Intentar eliminar el nombre del archivo del inicio
                if (archivoSeleccionado && textoRespuesta.includes(archivoSeleccionado)) {
                    const nombreSinExt = archivoSeleccionado.replace(/\.\w+$/, '');
                    if (textoRespuesta.includes(nombreSinExt)) {
                        console.log(`La respuesta contiene el nombre del archivo (${nombreSinExt}), intentando eliminarlo`);
                        const despuesDelNombre = textoRespuesta.indexOf(nombreSinExt) + nombreSinExt.length;
                        jsonLimpio = textoRespuesta.substring(despuesDelNombre);
                        
                        // Buscar nuevamente el inicio del JSON
                        const nuevoInicio = Math.min(
                            jsonLimpio.indexOf('{') >= 0 ? jsonLimpio.indexOf('{') : Number.MAX_SAFE_INTEGER,
                            jsonLimpio.indexOf('[') >= 0 ? jsonLimpio.indexOf('[') : Number.MAX_SAFE_INTEGER
                        );
                        
                        if (nuevoInicio < Number.MAX_SAFE_INTEGER) {
                            jsonLimpio = jsonLimpio.substring(nuevoInicio);
                        }
                    }
                }
                
                // Ahora intentar parsear el JSON limpio
                console.log("Intentando parsear:", jsonLimpio.substring(0, 50) + "...");
                const datos = JSON.parse(jsonLimpio);
                
                // Mostrar los datos en la tabla
                mostrarDatosEnTabla(datos);
                
            } catch (parseError) {
                console.error("Error al parsear JSON:", parseError);
                
                // Intento final: Si el archivo parece ser texto plano, crear un JSON manualmente
                try {
                    console.log("Intentando convertir texto plano a JSON");
                    
                    // Dividir por líneas y crear objetos
                    const lineas = textoRespuesta.split('\n').filter(l => l.trim());
                    
                    if (lineas.length > 0) {
                        // Verificar si la primera línea podría ser un encabezado
                        const primeraLinea = lineas[0];
                        const camposEncabezado = primeraLinea.split(/[\t,;|]/).map(c => c.trim());
                        
                        // Crear un array de objetos
                        const datos = [];
                        
                        for (let i = 1; i < lineas.length; i++) {
                            const valores = lineas[i].split(/[\t,;|]/);
                            const obj = {};
                            
                            // Asignar valores usando encabezados o índices
                            for (let j = 0; j < Math.max(camposEncabezado.length, valores.length); j++) {
                                const clave = j < camposEncabezado.length ? camposEncabezado[j] : `campo${j+1}`;
                                const valor = j < valores.length ? valores[j].trim() : '';
                                obj[clave] = valor;
                            }
                            
                            // Asegurar que tiene al menos un id
                            if (!obj.id) obj.id = `item-${i}`;
                            
                            datos.push(obj);
                        }
                        
                        mostrarDatosEnTabla(datos);
                        return;
                    }
                } catch (finalError) {
                    console.error("Todos los intentos de parseo fallaron:", finalError);
                    throw new Error("No se pudo procesar el contenido del archivo");
                }
                
                throw parseError;
            }
            
        } catch (error) {
            console.error('Error al cargar el archivo:', error);
            // alert('Error al cargar el archivo: ' + error.message);
        }
    });
}

// Función para mostrar los datos en la tabla
function mostrarDatosEnTabla(datos) {
    const tbody = document.getElementById('tabla-recortes');
    tbody.innerHTML = ''; // Limpiar tabla existente
    
    // Verificar si los datos son válidos
    if (!Array.isArray(datos)) {
        console.error('Datos no válidos:', datos);
        // Si datos es un objeto pero no un array, intentar ver si contiene un array
        if (datos && typeof datos === 'object') {
            // Buscar la primera propiedad que sea un array
            for (const prop in datos) {
                if (Array.isArray(datos[prop])) {
                    console.log(`Encontrado array en la propiedad '${prop}'`);
                    datos = datos[prop];
                    break;
                }
            }
        }
        
        // Si aún no es un array, intentar envolverlo
        if (!Array.isArray(datos)) {
            console.log('Intentando convertir a array');
            datos = [datos];
        }
    }
    
    // Llenar la tabla con los datos
    datos.forEach(item => {
        const fila = document.createElement('tr');
        
        // Formatear fecha y hora
        const fechaHora = `${item.fecha || ''} ${item.hora || ''}`.trim();
        
        // Crear celdas para cada campo
        fila.innerHTML = `
            <td>${item.autor || 'Sin autor'}</td>
            <td>${item.nombre ? item.nombre.replace(/\uFEFF|\u200B/g, '') : 'Sin nombre'}</td>
            <td class="contenido-celda">
                <div class="contenido-preview">${item.contenido || 'Sin contenido'}</div>
                ${item.pagina ? `<div class="pagina-info">Página ${item.pagina}</div>` : ''}
            </td>
            <td>${fechaHora || 'Sin fecha'}</td>
            <td class="acciones">
                <button class="btn-accion ver" data-id="${item.id || ''}">👁️</button>
                <button class="btn-accion etiquetar" data-id="${item.id || ''}">🏷️</button>
                ${item.visibilidad !== undefined ? 
                    (item.visibilidad ? 
                    '<span class="badge visible">Visible</span>' : 
                    '<span class="badge oculto">Oculto</span>') :
                    '<span class="badge">-</span>'}
            </td>
        `;
        
        tbody.appendChild(fila);
    });
    
    // Agregar event listeners a los botones
    document.querySelectorAll('.btn-accion.ver').forEach(btn => {
        btn.addEventListener('click', () => verDetalleCompleto(btn.dataset.id));
    });
    
    document.querySelectorAll('.btn-accion.etiquetar').forEach(btn => {
        btn.addEventListener('click', () => gestionarEtiquetas(btn.dataset.id));
    });
}

// Función para ver el detalle completo
function verDetalleCompleto(id) {
    console.log('Mostrar detalle completo del item:', id);
    // Aquí puedes implementar un modal o expandir la fila
}

// Función para gestionar etiquetas
function gestionarEtiquetas(id) {
    console.log('Gestionar etiquetas del item:', id);
    // Aquí puedes implementar un selector de etiquetas
}

// Inicialización cuando el DOM esté listo
document.addEventListener('DOMContentLoaded', function() {
    configurarSelectorArchivos();
    
    // Cargar lista de archivos (tu código existente)
    fetch('/mostrar-archivos')
        .then(response => response.json())
        .then(archivos => {
            const selector = document.getElementById('selector-archivos');
            archivos.forEach(archivo => {
                const option = document.createElement('option');
                option.value = archivo;
                option.textContent = archivo;
                selector.appendChild(option);
            });
        })
        .catch(error => console.error("Error al cargar archivos:", error));
});