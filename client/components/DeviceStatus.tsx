import { useState, useEffect } from 'react'

interface Device {
  id: string
  type: string
  location: string
  status: 'online' | 'offline'
  lastSeen: string
}

export default function DeviceStatus() {
  const [devices, setDevices] = useState<Device[]>([
    { id: 'device_001', type: 'temperature_sensor', location: 'warehouse_a', status: 'online', lastSeen: new Date().toISOString() },
    { id: 'device_002', type: 'humidity_sensor', location: 'warehouse_b', status: 'online', lastSeen: new Date().toISOString() },
    { id: 'device_003', type: 'motion_detector', location: 'office_floor_1', status: 'online', lastSeen: new Date().toISOString() },
    { id: 'device_004', type: 'camera', location: 'parking_lot', status: 'online', lastSeen: new Date().toISOString() },
    { id: 'device_005', type: 'controller', location: 'server_room', status: 'online', lastSeen: new Date().toISOString() },
  ])

  return (
    <div className="p-6">
      <div className="space-y-3">
        {devices.map((device) => (
          <div key={device.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
            <div>
              <h4 className="font-medium text-gray-900">{device.id}</h4>
              <p className="text-sm text-gray-600">{device.type.replace('_', ' ')}</p>
              <p className="text-xs text-gray-500">{device.location}</p>
            </div>
            <div className="text-right">
              <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                device.status === 'online' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
              }`}>
                {device.status}
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
