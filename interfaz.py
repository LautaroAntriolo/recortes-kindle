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
        self.root.title("Recortes Kindle - Visualizador")
        self.root.geometry("1000x750")
        
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
        filter_frame = ttk.Frame(self.root, padding="10")
        filter_frame.pack(fill=tk.X)
        
        ttk.Label(filter_frame, text="Filtrar por libro:").pack(side=tk.LEFT, padx=5)
        self.filter_entry = ttk.Entry(filter_frame, width=30)
        self.filter_entry.pack(side=tk.LEFT, padx=5)
        ttk.Button(filter_frame, text="Aplicar filtro", command=self.aplicar_filtro).pack(side=tk.LEFT, padx=5)
        ttk.Button(filter_frame, text="Limpiar filtro", command=self.limpiar_filtro).pack(side=tk.LEFT, padx=5)
        
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
        nav_frame = ttk.Frame(right_frame, padding="10")
        nav_frame.pack(fill=tk.X, pady=10)
        
        ttk.Button(nav_frame, text="Anterior", command=self.anterior_registro).pack(side=tk.LEFT, padx=5)
        ttk.Button(nav_frame, text="Siguiente", command=self.siguiente_registro).pack(side=tk.LEFT, padx=5)
        ttk.Button(nav_frame, text="Guardar cambios", command=self.guardar_cambios).pack(side=tk.RIGHT, padx=5)
        ttk.Button(nav_frame, text="Exportar JSON", command=self.exportar_json).pack(side=tk.RIGHT, padx=5)
    
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
        """Guarda los cambios en el registro actual"""
        if not self.registros:
            return
            
        try:
            # Actualizar el registro con los valores editados
            self.registros[self.indice_actual] = {
                "id": int(self.id_var.get()),
                "autor": self.autor_var.get(),
                "nombre": self.libro_var.get(),
                "pagina": int(self.pagina_var.get()),
                "contenido": self.contenido_text.get("1.0", tk.END).strip(),
                "visibilidad": self.visibilidad_var.get(),
                "fecha": self.fecha_var.get(),
                "hora": self.hora_var.get()
            }
            
            # Actualizar la lista ordenada
            self.registros_ordenados_por_id = sorted(self.registros, key=lambda x: x["id"])
            
            # Actualizar la vista
            self.actualizar_lista_registros()
            messagebox.showinfo("Éxito", "Cambios guardados en memoria")
            
        except ValueError as e:
            messagebox.showerror("Error", f"Datos inválidos: {str(e)}")
    
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
    
    def aplicar_filtro(self):
        """Aplica un filtro avanzado buscando en múltiples campos"""
        palabra_busqueda = self.filter_entry.get().strip().lower()
        self.filtro_actual = palabra_busqueda
        
        if not palabra_busqueda:
            self.limpiar_filtro()
            return
        
        # Limpiar el Treeview actual
        for item in self.tree.get_children():
            self.tree.delete(item)
    
        # Función auxiliar para buscar en múltiples campos
        def coincide(registro):
            campos_a_buscar = [
                registro["nombre"].lower(),
                registro["autor"].lower(),
                registro["contenido"].lower(),
                str(registro["pagina"])
            ]
            return any(palabra_busqueda in campo for campo in campos_a_buscar)
    
        # Filtrar registros
        registros_filtrados = [reg for reg in self.registros if coincide(reg)]
        
        # Llenar el Treeview con los datos filtrados
        for reg in registros_filtrados:
            visible = "Sí" if reg["visibilidad"] else "No"
            self.tree.insert("", tk.END, values=(
                reg["id"],
                reg["nombre"],
                reg["autor"],
                reg["pagina"],
                visible
            ), iid=str(reg["id"]))
        
        # Manejar selección y feedback al usuario
        if registros_filtrados:
            self.indice_actual = 0
            self.tree.selection_set(str(registros_filtrados[0]["id"]))
            self.mostrar_registro(0)
            self.root.title(f"Recortes Kindle - {len(registros_filtrados)} resultados encontrados")
        else:
            messagebox.showinfo("Información", "No se encontraron coincidencias")
            self.root.title("Recortes Kindle - 0 resultados encontrados")
    
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

if __name__ == "__main__":
    root = tk.Tk()
    app = RecortesKindleApp(root)
    root.mainloop()