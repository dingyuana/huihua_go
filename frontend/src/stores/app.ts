import { defineStore } from 'pinia'

interface AppState {
  currentPeriod: {
    year: number
    month: number
  }
  sidebarCollapsed: boolean
}

export const useAppStore = defineStore('app', {
  state: (): AppState => ({
    currentPeriod: {
      year: new Date().getFullYear(),
      month: new Date().getMonth() + 1
    },
    sidebarCollapsed: false
  }),
  
  actions: {
    setCurrentPeriod(year: number, month: number) {
      this.currentPeriod = { year, month }
    },
    
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
    }
  }
})