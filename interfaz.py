import tkinter as tk
from tkinter import filedialog, scrolledtext, messagebox, ttk
import subprocess
import json
import os
import sys
import tempfile

class RecortesKindleApp:
    def __init__(self, root):
        self.root = root
        self.root.title("Visualizador de notas")
        self.root.geometry("1000x750")
        
        # Lista para seguir archivos temporales
        self.archivos_temporales = []
        
        # Registrar el evento de cierre
        self.root.protocol("WM_DELETE_WINDOW", self.on_close)
        
        # Configurar el tema y estilo
        self.style = ttk.Style()
        self.style.theme_use("clam")
        
        # Variables
        self.txt_file_path = ""
        self.registros = []
        self.indice_actual = 0
        self.filtro_actual = ""
        self.registros_ordenados_por_id = []
        
        # Crear la interfaz
        self.create_widgets()
    
    def create_widgets(self):
        # Frame superior para selección de archivo
        top_frame = ttk.Frame(self.root, padding="10")
        top_frame.pack(fill=tk.X)
        
        ttk.Label(top_frame, text="Archivo de recortes:").pack(side=tk.LEFT, padx=5)
        
        self.file_var = tk.StringVar(value="Ningún archivo seleccionado")
        ttk.Label(top_frame, textvariable=self.file_var, width=40).pack(side=tk.LEFT, padx=5)
        
        ttk.Button(top_frame, text="Seleccionar archivo", command=self.select_file).pack(side=tk.LEFT, padx=5)
        ttk.Button(top_frame, text="Procesar", command=self.process_file).pack(side=tk.LEFT, padx=5)
        
        # Frame de filtros
        filter_frame = ttk.LabelFrame(self.root, text="Búsqueda", padding="10")
        filter_frame.pack(fill=tk.X, padx=10, pady=5)

        # Pestañas para diferentes tipos de búsqueda
        search_notebook = ttk.Notebook(filter_frame)
        search_notebook.pack(fill=tk.X, expand=True)

        # Pestaña de búsqueda por libro
        book_frame = ttk.Frame(search_notebook, padding=5)
        search_notebook.add(book_frame, text="Por Libro")

        ttk.Label(book_frame, text="Nombre del libro:").pack(side=tk.LEFT, padx=5)
        self.book_search_entry = ttk.Entry(book_frame, width=30)
        self.book_search_entry.pack(side=tk.LEFT, padx=5)
        ttk.Button(book_frame, text="Buscar", command=self.buscar_por_libro).pack(side=tk.LEFT, padx=5)

        # Pestaña de búsqueda por autor
        author_frame = ttk.Frame(search_notebook, padding=5)
        search_notebook.add(author_frame, text="Por Autor")

        ttk.Label(author_frame, text="Nombre del autor:").pack(side=tk.LEFT, padx=5)
        self.author_search_entry = ttk.Entry(author_frame, width=30)
        self.author_search_entry.pack(side=tk.LEFT, padx=5)
        ttk.Button(author_frame, text="Buscar", command=self.buscar_por_autor).pack(side=tk.LEFT, padx=5)

        # Pestaña de búsqueda por contenido
        content_frame = ttk.Frame(search_notebook, padding=5)
        search_notebook.add(content_frame, text="Por Contenido")

        ttk.Label(content_frame, text="Texto en contenido:").pack(side=tk.LEFT, padx=5)
        self.content_search_entry = ttk.Entry(content_frame, width=30)
        self.content_search_entry.pack(side=tk.LEFT, padx=5)
        ttk.Button(content_frame, text="Buscar", command=self.buscar_por_contenido).pack(side=tk.LEFT, padx=5)
        ttk.Button(content_frame, text="Exportar JSON", command=self.exportar_resultados_busqueda).pack(side=tk.LEFT, padx=5)
        
        # Frame principal dividido en dos
        main_frame = ttk.Frame(self.root, padding="10")
        main_frame.pack(fill=tk.BOTH, expand=True)
        
        # Panel izquierdo: Lista de registros
        left_frame = ttk.LabelFrame(main_frame, text="Registros", padding="10")
        left_frame.pack(side=tk.LEFT, fill=tk.BOTH, expand=True, padx=(0, 5))
        
        # Crear un Treeview con columnas
        self.tree = ttk.Treeview(left_frame, columns=("id", "libro", "autor", "pagina", "visible"), show="headings")
        self.tree.heading("id", text="ID", command=lambda: self.ordenar_por("id"))
        self.tree.heading("libro", text="Libro", command=lambda: self.ordenar_por("nombre"))
        self.tree.heading("autor", text="Autor", command=lambda: self.ordenar_por("autor"))
        self.tree.heading("pagina", text="Página", command=lambda: self.ordenar_por("pagina"))
        self.tree.heading("visible", text="Visible")
        
        self.tree.column("id", width=50, anchor=tk.CENTER)
        self.tree.column("libro", width=150)
        self.tree.column("autor", width=150)
        self.tree.column("pagina", width=60, anchor=tk.CENTER)
        self.tree.column("visible", width=50, anchor=tk.CENTER)
        
        self.tree.pack(side=tk.LEFT, fill=tk.BOTH, expand=True)
        
        # Scrollbar para el Treeview
        scrollbar = ttk.Scrollbar(left_frame, orient=tk.VERTICAL, command=self.tree.yview)
        scrollbar.pack(side=tk.RIGHT, fill=tk.Y)
        self.tree.configure(yscrollcommand=scrollbar.set)
        self.tree.bind("<<TreeviewSelect>>", self.on_registro_select)
        
        # Panel derecho: Detalles del registro seleccionado
        right_frame = ttk.LabelFrame(main_frame, text="Detalles", padding="10")
        right_frame.pack(side=tk.RIGHT, fill=tk.BOTH, expand=True, padx=(5, 0))
        
        # Detalles del registro
        details_frame = ttk.Frame(right_frame)
        details_frame.pack(fill=tk.X, pady=(0, 10))
        
        # Campos editables
        ttk.Label(details_frame, text="ID:").grid(row=0, column=0, sticky=tk.W, pady=2)
        self.id_var = tk.StringVar()
        ttk.Label(details_frame, textvariable=self.id_var, width=10).grid(row=0, column=1, sticky=tk.W, pady=2)
        
        ttk.Label(details_frame, text="Autor:").grid(row=1, column=0, sticky=tk.W, pady=2)
        self.autor_var = tk.StringVar()
        ttk.Entry(details_frame, textvariable=self.autor_var, width=50).grid(row=1, column=1, sticky=tk.W, pady=2)
        
        ttk.Label(details_frame, text="Libro:").grid(row=2, column=0, sticky=tk.W, pady=2)
        self.libro_var = tk.StringVar()
        ttk.Entry(details_frame, textvariable=self.libro_var, width=50).grid(row=2, column=1, sticky=tk.W, pady=2)
        
        ttk.Label(details_frame, text="Página:").grid(row=3, column=0, sticky=tk.W, pady=2)
        self.pagina_var = tk.StringVar()
        ttk.Entry(details_frame, textvariable=self.pagina_var, width=10).grid(row=3, column=1, sticky=tk.W, pady=2)
        
        ttk.Label(details_frame, text="Fecha:").grid(row=4, column=0, sticky=tk.W, pady=2)
        self.fecha_var = tk.StringVar()
        ttk.Entry(details_frame, textvariable=self.fecha_var, width=20).grid(row=4, column=1, sticky=tk.W, pady=2)
        
        ttk.Label(details_frame, text="Hora:").grid(row=5, column=0, sticky=tk.W, pady=2)
        self.hora_var = tk.StringVar()
        ttk.Entry(details_frame, textvariable=self.hora_var, width=20).grid(row=5, column=1, sticky=tk.W, pady=2)
        
        # Visibilidad
        self.visibilidad_var = tk.BooleanVar()
        ttk.Checkbutton(details_frame, text="Visible", variable=self.visibilidad_var).grid(row=6, column=0, columnspan=2, sticky=tk.W, pady=5)
        
        # Contenido del recorte
        ttk.Label(right_frame, text="Contenido:").pack(anchor=tk.W)
        self.contenido_text = scrolledtext.ScrolledText(right_frame, wrap=tk.WORD, height=12)
        self.contenido_text.pack(fill=tk.BOTH, expand=True)
        
        # Frame inferior para botones de navegación
        nav_frame = ttk.Frame(right_frame, padding="5")
        nav_frame.pack(fill=tk.X, pady=5)
        
        ttk.Button(nav_frame, text="Anterior", command=self.anterior_registro).pack(side=tk.LEFT, padx=5)
        ttk.Button(nav_frame, text="Siguiente", command=self.siguiente_registro).pack(side=tk.LEFT, padx=5)
        ttk.Button(nav_frame, text="Guardar cambios", command=self.guardar_cambios).pack(side=tk.RIGHT, padx=5)
        ttk.Button(nav_frame, text="Exportar JSON", command=self.exportar_json).pack(side=tk.RIGHT, padx=5)
        ttk.Button(nav_frame, text="Eliminar", command=self.eliminar_registro, style='Danger.TButton').pack(side=tk.LEFT, padx=5)
        
        self.style.configure('Danger.TButton', foreground='white', background='#dc3545')
        self.style.map('Danger.TButton', background=[('active', '#c82333'), ('pressed', '#bd2130')])

    def eliminar_registro(self):
        """Elimina el registro actualmente seleccionado"""
        if not self.registros:
            messagebox.showwarning("Advertencia", "No hay registros para eliminar")
            return
        
        # Confirmar con el usuario
        reg_actual = self.registros[self.indice_actual]
        confirmacion = messagebox.askyesno(
            "Confirmar eliminación",
            f"¿Estás seguro de que deseas eliminar este registro?\n\n"
            f"Libro: {reg_actual['nombre']}\n"
            f"Autor: {reg_actual['autor']}\n"
            f"ID: {reg_actual['id']}"
        )
        
        if not confirmacion:
            return
        
        try:
            # Eliminar el registro de la lista
            registro_eliminado = self.registros.pop(self.indice_actual)
            
            # Actualizar la lista ordenada por ID
            self.registros_ordenados_por_id = [r for r in self.registros_ordenados_por_id 
                                            if r['id'] != registro_eliminado['id']]
            
            # Actualizar el archivo TXT si existe
            if self.txt_file_path:
                self.actualizar_archivo_txt()
            
            # Actualizar la interfaz
            if self.registros:
                # Ajustar el índice actual para no salirse de los límites
                if self.indice_actual >= len(self.registros):
                    self.indice_actual = len(self.registros) - 1
                
                self.actualizar_lista_registros()
                self.tree.selection_set(str(self.registros[self.indice_actual]['id']))
                self.mostrar_registro(self.indice_actual)
            else:
                # No quedan registros
                self.limpiar_interfaz()
            
            messagebox.showinfo("Éxito", "Registro eliminado correctamente")
            
        except Exception as e:
            messagebox.showerror("Error", f"No se pudo eliminar el registro: {str(e)}")
            # Revertir los cambios en memoria si falla
            if 'registro_eliminado' in locals():
                self.registros.insert(self.indice_actual, registro_eliminado)
                self.registros_ordenados_por_id = sorted(self.registros, key=lambda x: x["id"]) 
    
    def select_file(self):
        """Abre diálogo para seleccionar archivo de recortes"""
        file_path = filedialog.askopenfilename(
            title="Seleccionar archivo de recortes",
            filetypes=[("Archivos de texto", "*.txt"), ("Todos los archivos", "*.*")]
        )
        
        if file_path:
            self.txt_file_path = file_path
            self.file_var.set(os.path.basename(file_path))
    
    def process_file(self):
        """Procesa el archivo seleccionado usando el programa Go"""
        if not self.txt_file_path:
            messagebox.showwarning("Advertencia", "Primero seleccione un archivo TXT")
            return
        
        try:
            # Determinar la ruta del ejecutable Go
            go_executable = "./recortesKindle"
            if sys.platform == "win32":
                go_executable += ".exe"
            
            # Verificar si el ejecutable existe
            if not os.path.exists(go_executable):
                messagebox.showerror("Error", 
                    f"No se encontró el ejecutable Go en: {os.path.abspath(go_executable)}\n"
                    f"Por favor, compile su programa Go primero con 'go build'")
                return
            
            # Usar un archivo temporal para el JSON
            with tempfile.NamedTemporaryFile(delete=False, suffix=".json") as temp_file:
                temp_json_path = temp_file.name
            
            # Ejecutar el programa Go
            process = subprocess.run(
                [go_executable, self.txt_file_path, temp_json_path],
                capture_output=True,
                text=True,
                encoding='utf-8',
                check=True
            )
            
            try:
                # Intentar parsear el JSON desde la salida estándar
                self.registros = json.loads(process.stdout)
                self.registros_ordenados_por_id = sorted(self.registros, key=lambda x: x["id"])
                self.actualizar_lista_registros()
                messagebox.showinfo("Éxito", f"Procesados {len(self.registros)} recortes")
            except json.JSONDecodeError:
                # Si falla, intentar leer del archivo temporal
                try:
                    with open(temp_json_path, 'r', encoding='utf-8') as f:
                        self.registros = json.load(f)
                    self.registros_ordenados_por_id = sorted(self.registros, key=lambda x: x["id"])
                    self.actualizar_lista_registros()
                    messagebox.showinfo("Éxito", f"Procesados {len(self.registros)} recortes (leído de archivo temporal)")
                except Exception as e:
                    messagebox.showerror("Error", f"No se pudo leer el JSON: {str(e)}")
            
            # Eliminar el archivo temporal
            try:
                os.unlink(temp_json_path)
            except:
                pass
                
        except subprocess.CalledProcessError as e:
            messagebox.showerror("Error", f"Error al procesar el archivo:\n{e.stderr}")
        except Exception as e:
            messagebox.showerror("Error", f"Error inesperado: {str(e)}")
    
    def actualizar_lista_registros(self):
        """Actualiza el Treeview con los registros actuales"""
        # Limpiar el Treeview actual
        for item in self.tree.get_children():
            self.tree.delete(item)
        
        # Llenar el Treeview con los datos
        for reg in self.registros:
            visible = "Sí" if reg["visibilidad"] else "No"
            self.tree.insert("", tk.END, values=(
                reg["id"],
                reg["nombre"],
                reg["autor"],
                reg["pagina"],
                visible
            ), iid=str(reg["id"]))
        
        # Seleccionar el primer registro si hay alguno
        if self.registros:
            self.indice_actual = 0
            self.tree.selection_set(str(self.registros[0]["id"]))
            self.mostrar_registro(0)
    
    def mostrar_registro(self, indice):
        """Muestra los detalles del registro seleccionado"""
        if not self.registros or indice < 0 or indice >= len(self.registros):
            return
        
        reg = self.registros[indice]
        self.id_var.set(str(reg["id"]))
        self.autor_var.set(reg["autor"])
        self.libro_var.set(reg["nombre"])
        self.pagina_var.set(str(reg["pagina"]))
        self.fecha_var.set(reg["fecha"])
        self.hora_var.set(reg["hora"])
        self.visibilidad_var.set(reg["visibilidad"])
        
        # Actualizar el contenido
        self.contenido_text.delete(1.0, tk.END)
        self.contenido_text.insert(tk.END, reg["contenido"])
    
    def on_registro_select(self, event):
        """Maneja la selección de un registro en el Treeview"""
        selection = self.tree.selection()
        if selection:
            selected_id = int(selection[0])
            for i, reg in enumerate(self.registros):
                if reg["id"] == selected_id:
                    self.indice_actual = i
                    self.mostrar_registro(i)
                    break
    
    def siguiente_registro(self):
        """Navega al siguiente registro por ID"""
        if not self.registros:
            return
        
        current_id = self.registros[self.indice_actual]["id"]
        
        # Encontrar la posición actual en la lista ordenada por ID
        ids_ordenados = [r["id"] for r in self.registros_ordenados_por_id]
        current_pos = ids_ordenados.index(current_id)
        
        if current_pos < len(self.registros_ordenados_por_id) - 1:
            next_id = self.registros_ordenados_por_id[current_pos + 1]["id"]
            
            # Buscar el índice en la lista original
            for i, reg in enumerate(self.registros):
                if reg["id"] == next_id:
                    self.indice_actual = i
                    self.tree.selection_set(str(next_id))
                    self.tree.focus(str(next_id))
                    self.tree.see(str(next_id))
                    self.mostrar_registro(i)
                    break
    
    def anterior_registro(self):
        """Navega al registro anterior por ID"""
        if not self.registros:
            return
        
        current_id = self.registros[self.indice_actual]["id"]
        
        # Encontrar la posición actual en la lista ordenada por ID
        ids_ordenados = [r["id"] for r in self.registros_ordenados_por_id]
        current_pos = ids_ordenados.index(current_id)
        
        if current_pos > 0:
            prev_id = self.registros_ordenados_por_id[current_pos - 1]["id"]
            
            # Buscar el índice en la lista original
            for i, reg in enumerate(self.registros):
                if reg["id"] == prev_id:
                    self.indice_actual = i
                    self.tree.selection_set(str(prev_id))
                    self.tree.focus(str(prev_id))
                    self.tree.see(str(prev_id))
                    self.mostrar_registro(i)
                    break
    
    def guardar_cambios(self):
        """Guarda los cambios en el registro actual y actualiza el archivo TXT"""
        if not self.registros:
            return
            
        try:
            # Actualizar el registro con los valores editados
            reg_actualizado = {
                "id": int(self.id_var.get()),
                "autor": self.autor_var.get(),
                "nombre": self.libro_var.get(),
                "pagina": int(self.pagina_var.get()),
                "contenido": self.contenido_text.get("1.0", tk.END).strip(),
                "visibilidad": self.visibilidad_var.get(),
                "fecha": self.fecha_var.get(),
                "hora": self.hora_var.get()
            }
            
            self.registros[self.indice_actual] = reg_actualizado
            self.registros_ordenados_por_id = sorted(self.registros, key=lambda x: x["id"])
            
            # Actualizar el archivo TXT original
            if self.txt_file_path:
                self.actualizar_archivo_txt()
            
            # Actualizar la vista
            self.actualizar_lista_registros()
            messagebox.showinfo("Éxito", "Cambios guardados en memoria y archivo TXT")
            
        except ValueError as e:
            messagebox.showerror("Error", f"Datos inválidos: {str(e)}")

    def actualizar_archivo_txt(self):
        """Reconstruye y guarda el archivo TXT original con los cambios"""
        try:
            # Ordenar los registros como en el archivo original
            registros_ordenados = sorted(self.registros, key=lambda x: x["id"])
            
            lineas = []
            for reg in registros_ordenados:
                # Solo incluir registros visibles si así se desea
                if reg["visibilidad"]:
                    linea = "==========\n"
                    linea += f"{reg['nombre']} ({reg['autor']})\n"
                    
                    # Construir la línea de metadatos de manera dinámica
                    meta_line = f"- Tu recorte en la página {reg['pagina']}"
                    
                    # Añadir posición solo si existe
                    if reg.get('posicion', ''):
                        meta_line += f" | Posición {reg['posicion']}"
                    
                    # Añadir fecha y hora
                    meta_line += f" | Añadido el {reg['fecha']} a las {reg['hora']}\n\n"
                    linea += meta_line
                    
                    # Añadir contenido
                    linea += f"{reg['contenido']}\n\n"
                    
                    lineas.append(linea)
            
            # Escribir el archivo completo
            with open(self.txt_file_path, 'w', encoding='utf-8') as f:
                f.writelines(lineas)
                
        except Exception as e:
            messagebox.showerror("Error", f"No se pudo guardar el archivo TXT: {str(e)}")
            
    def exportar_json(self):
        """Exporta los registros a un archivo JSON"""
        if not self.registros:
            messagebox.showwarning("Advertencia", "No hay datos para exportar")
            return
            
        file_path = filedialog.asksaveasfilename(
            defaultextension=".json",
            filetypes=[("Archivos JSON", "*.json"), ("Todos los archivos", "*.*")]
        )
        
        if file_path:
            try:
                with open(file_path, 'w', encoding='utf-8') as f:
                    json.dump(self.registros, f, indent=2, ensure_ascii=False)
                messagebox.showinfo("Éxito", f"Datos exportados a {file_path}")
            except Exception as e:
                messagebox.showerror("Error", f"No se pudo exportar: {str(e)}")
    
    def buscar_por_libro(self):
        """Filtra registros por nombre de libro"""
        termino = self.book_search_entry.get().lower()
        if not termino:
            self.actualizar_lista_registros()
            return
        
        resultados = [reg for reg in self.registros if termino in reg["nombre"].lower()]
        self.mostrar_resultados_busqueda(resultados)

    def buscar_por_autor(self):
        """Filtra registros por autor"""
        termino = self.author_search_entry.get().lower()
        if not termino:
            self.actualizar_lista_registros()
            return
        
        resultados = [reg for reg in self.registros if termino in reg["autor"].lower()]
        self.mostrar_resultados_busqueda(resultados)

    def buscar_por_contenido(self):
        """Filtra registros por contenido"""
        termino = self.content_search_entry.get().lower()
        if not termino:
            self.actualizar_lista_registros()
            return
        
        resultados = [reg for reg in self.registros if termino in reg["contenido"].lower()]
        self.mostrar_resultados_busqueda(resultados)
        self.resultados_busqueda_actual = resultados  # Guardar para posible exportación

    def mostrar_resultados_busqueda(self, resultados):
        """Muestra los resultados de búsqueda en el Treeview"""
        self.tree.delete(*self.tree.get_children())
        
        for reg in resultados:
            visible = "Sí" if reg["visibilidad"] else "No"
            self.tree.insert("", tk.END, values=(
                reg["id"],
                reg["nombre"],
                reg["autor"],
                reg["pagina"],
                visible
            ), iid=str(reg["id"]))
        
        if resultados:
            self.tree.selection_set(str(resultados[0]["id"]))
            self.mostrar_registro(0)
    
    def exportar_resultados_busqueda(self):
        """Exporta los resultados de la última búsqueda a un archivo JSON temporal"""
        if not hasattr(self, 'resultados_busqueda_actual') or not self.resultados_busqueda_actual:
            messagebox.showwarning("Advertencia", "No hay resultados de búsqueda para exportar")
            return
        
        try:
            # Crear un archivo temporal que se eliminará cuando se cierre
            with tempfile.NamedTemporaryFile(
                prefix="kindle_search_", 
                suffix=".json", 
                delete=False,  # No borrar inmediatamente para poder usarlo
                mode='w'
            ) as temp_file:
                json.dump(self.resultados_busqueda_actual, temp_file, indent=2, ensure_ascii=False)
                temp_path = temp_file.name
            
            # Registrar este archivo para eliminación al salir
            self.archivos_temporales = getattr(self, 'archivos_temporales', [])
            self.archivos_temporales.append(temp_path)
            
            messagebox.showinfo("Éxito", 
                            f"Resultados exportados a archivo temporal:\n{temp_path}\n\n"
                            f"Este archivo se eliminará cuando cierres la aplicación.")
            
            # Devuelve la ruta para que puedas usar este archivo en otras partes de tu app
            return temp_path
            
        except Exception as e:
            messagebox.showerror("Error", f"No se pudo exportar el JSON: {str(e)}")
            return None
    
    def limpiar_filtro(self):
        """Limpia el filtro aplicado"""
        self.filtro_actual = ""
        self.filter_entry.delete(0, tk.END)
        self.actualizar_lista_registros()
    
    def ordenar_por(self, campo):
        """Ordena los registros por el campo especificado"""
        if not self.registros:
            return
            
        # Ordenar los registros
        self.registros.sort(key=lambda x: str(x[campo]).lower() if isinstance(x[campo], str) else x[campo])
        
        # Actualizar la vista
        self.actualizar_lista_registros()

    def on_close(self):
        """Limpia los archivos temporales y cierra la aplicación"""
        # Eliminar archivos temporales
        for archivo in self.archivos_temporales:
            try:
                if os.path.exists(archivo):
                    os.unlink(archivo)
            except Exception:
                pass  # Ignorar errores al eliminar
        
        # Cerrar la aplicación
        self.root.destroy()
            
if __name__ == "__main__":
    root = tk.Tk()
    app = RecortesKindleApp(root)
    root.mainloop()