import React, { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
  Button,
} from '../ui/Dialog';

interface TimePickerDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  currentTime: Date;
  onTimeSelect: (time: Date) => void;
}

export function TimePickerDialog({
  open,
  onOpenChange,
  currentTime,
  onTimeSelect,
}: TimePickerDialogProps) {
  const [selectedDate, setSelectedDate] = useState(
    currentTime.toISOString().slice(0, 10)
  );
  const [hours, setHours] = useState(
    currentTime.getHours().toString().padStart(2, '0')
  );
  const [minutes, setMinutes] = useState(
    currentTime.getMinutes().toString().padStart(2, '0')
  );
  const [seconds, setSeconds] = useState(
    currentTime.getSeconds().toString().padStart(2, '0')
  );

  const handleQuickSelect = (option: string) => {
    const now = new Date();
    let targetTime: Date;

    switch (option) {
      case '1hr':
        targetTime = new Date(now.getTime() - 60 * 60 * 1000);
        break;
      case '6hr':
        targetTime = new Date(now.getTime() - 6 * 60 * 60 * 1000);
        break;
      case 'today':
        targetTime = new Date(now.setHours(0, 0, 0, 0));
        break;
      case 'yesterday':
        targetTime = new Date(now.setDate(now.getDate() - 1));
        targetTime.setHours(0, 0, 0, 0);
        break;
      default:
        return;
    }

    setSelectedDate(targetTime.toISOString().slice(0, 10));
    setHours(targetTime.getHours().toString().padStart(2, '0'));
    setMinutes(targetTime.getMinutes().toString().padStart(2, '0'));
    setSeconds(targetTime.getSeconds().toString().padStart(2, '0'));
  };

  const handleGo = () => {
    const dateTime = new Date(`${selectedDate}T${hours}:${minutes}:${seconds}`);
    if (!isNaN(dateTime.getTime())) {
      onTimeSelect(dateTime);
      onOpenChange(false);
    }
  };

  const handleCancel = () => {
    // Reset to current time
    setSelectedDate(currentTime.toISOString().slice(0, 10));
    setHours(currentTime.getHours().toString().padStart(2, '0'));
    setMinutes(currentTime.getMinutes().toString().padStart(2, '0'));
    setSeconds(currentTime.getSeconds().toString().padStart(2, '0'));
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent onClose={handleCancel} className="max-w-md">
        <DialogHeader>
          <DialogTitle>Select Playback Time</DialogTitle>
        </DialogHeader>

        <DialogBody className="space-y-4">
          {/* Date Picker */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Date
            </label>
            <input
              type="date"
              value={selectedDate}
              onChange={(e) => setSelectedDate(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          {/* Time Picker */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Time (HH:MM:SS)
            </label>
            <div className="flex items-center gap-2">
              <input
                type="number"
                min="0"
                max="23"
                value={hours}
                onChange={(e) => setHours(e.target.value.padStart(2, '0'))}
                className="w-20 px-3 py-2 border border-gray-300 rounded-lg text-center font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="HH"
              />
              <span className="text-gray-500 font-bold">:</span>
              <input
                type="number"
                min="0"
                max="59"
                value={minutes}
                onChange={(e) => setMinutes(e.target.value.padStart(2, '0'))}
                className="w-20 px-3 py-2 border border-gray-300 rounded-lg text-center font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="MM"
              />
              <span className="text-gray-500 font-bold">:</span>
              <input
                type="number"
                min="0"
                max="59"
                value={seconds}
                onChange={(e) => setSeconds(e.target.value.padStart(2, '0'))}
                className="w-20 px-3 py-2 border border-gray-300 rounded-lg text-center font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="SS"
              />
            </div>
          </div>

          {/* Quick Select */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Quick Select
            </label>
            <div className="grid grid-cols-2 gap-2">
              <button
                onClick={() => handleQuickSelect('1hr')}
                className="px-3 py-2 bg-gray-100 hover:bg-gray-200 rounded-lg text-sm font-medium transition-colors"
              >
                1 Hour Ago
              </button>
              <button
                onClick={() => handleQuickSelect('6hr')}
                className="px-3 py-2 bg-gray-100 hover:bg-gray-200 rounded-lg text-sm font-medium transition-colors"
              >
                6 Hours Ago
              </button>
              <button
                onClick={() => handleQuickSelect('today')}
                className="px-3 py-2 bg-gray-100 hover:bg-gray-200 rounded-lg text-sm font-medium transition-colors"
              >
                Today 00:00
              </button>
              <button
                onClick={() => handleQuickSelect('yesterday')}
                className="px-3 py-2 bg-gray-100 hover:bg-gray-200 rounded-lg text-sm font-medium transition-colors"
              >
                Yesterday
              </button>
            </div>
          </div>

          {/* Preview */}
          <div className="p-3 bg-blue-50 border border-blue-200 rounded-lg">
            <p className="text-sm text-gray-700">
              <span className="font-medium">Selected:</span>{' '}
              {new Date(`${selectedDate}T${hours}:${minutes}:${seconds}`).toLocaleString('en-US', {
                month: 'short',
                day: 'numeric',
                year: 'numeric',
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit',
                hour12: false,
              })}
            </p>
          </div>
        </DialogBody>

        <DialogFooter>
          <Button variant="secondary" onClick={handleCancel}>
            Cancel
          </Button>
          <Button variant="primary" onClick={handleGo}>
            Go to Time
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
