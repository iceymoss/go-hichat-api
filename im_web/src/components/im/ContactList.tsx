'use client';

import React, { useState, useMemo, useRef, useCallback } from 'react';
import { contacts, contactGroups, getAlphabetIndex, type Contact } from '@/lib/mock-data';
import { useIMStore } from '@/lib/im-store';
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar';
import { Input } from '@/components/ui/input';
import { Search, UserPlus, Users, Tag, Megaphone, ChevronRight } from 'lucide-react';
// Contact detail is shown via ContactDetailPanel in IMLayout right panel

const iconMap: Record<string, React.ReactNode> = {
  UserPlus: <UserPlus className="w-[22px] h-[22px]" />,
  Users: <Users className="w-[22px] h-[22px]" />,
  Tag: <Tag className="w-[22px] h-[22px]" />,
  Megaphone: <Megaphone className="w-[22px] h-[22px]" />,
};

export default function ContactList() {
  const [searchQuery, setSearchQuery] = useState('');
  const [backendResults, setBackendResults] = useState<Contact[]>([]);
  const { setSelectedContactId, selectedContactId, setShowFriendRequests, friendRequestUnreadCount, setShowGroupPanel, groupAppUnreadCount, currentUser } = useIMStore();
  const scrollRef = useRef<HTMLDivElement>(null);

  // Search backend when query looks like phone/email or after debounce
  React.useEffect(() => {
    const q = searchQuery.trim();
    if (!q || !currentUser?.token) { setBackendResults([]); return; }
    const isPhone = /^1[3-9]\d{2,10}$/.test(q);
    const isEmail = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(q);
    if (!isPhone && !isEmail) { setBackendResults([]); return; }
    const params = isPhone ? `phone=${q}` : `email=${q}`;
    const timer = setTimeout(() => {
      fetch(`/api/user/search?${params}`, { headers: { Authorization: `Bearer ${currentUser.token}` } })
        .then(r => r.json())
        .then(d => {
          if (d.success && d.data?.users?.length) {
            setBackendResults(d.data.users.map((u: any) => ({
              id: u.id, name: u.nickname, avatar: u.avatar || '',
              phone: u.mobile, region: u.region, signature: u.introduction,
              pinyin: '', letter: '#', online: false, account: u.id,
            })));
          } else setBackendResults([]);
        })
        .catch(() => setBackendResults([]));
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery, currentUser?.token]);

  const filteredContacts = useMemo(() => {
    if (!searchQuery.trim()) return contacts;
    const q = searchQuery.toLowerCase();
    const local = contacts.filter(c =>
      c.name.toLowerCase().includes(q) || c.pinyin.toLowerCase().includes(q)
    );
    // Merge backend results, avoid duplicates
    const localIds = new Set(local.map(c => c.id));
    return [...local, ...backendResults.filter(c => !localIds.has(c.id))];
  }, [searchQuery, backendResults]);

  // Group contacts by letter
  const groupedContacts = useMemo(() => {
    const groups: Record<string, Contact[]> = {};
    filteredContacts.forEach(c => {
      if (!groups[c.letter]) groups[c.letter] = [];
      groups[c.letter].push(c);
    });
    return Object.entries(groups).sort(([a], [b]) => a.localeCompare(b));
  }, [filteredContacts]);

  const alphabetIndex = useMemo(() => {
    return getAlphabetIndex();
  }, []);

  const scrollToLetter = useCallback((letter: string) => {
    const el = document.getElementById(`letter-${letter}`);
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
  }, []);

  const totalContacts = contacts.length;

  const handleContactClick = useCallback((contact: Contact) => {
    setSelectedContactId(contact.id);
  }, [setSelectedContactId]);

  return (
    <div className="h-full flex flex-col">
      {/* Search Bar — TG rounded, subtle bg, no border */}
      <div className="px-3 py-2.5 shrink-0">
        <div className="relative">
          <Search
            className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4"
            style={{ color: '#A2ACB5' }}
          />
          <Input
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="搜索"
            style={{
              backgroundColor: '#F0F2F5',
              border: 'none',
              borderRadius: '22px',
              height: '36px',
              paddingLeft: '36px',
              fontSize: '13px',
              color: '#1C2733',
              boxShadow: 'none',
            }}
            className="focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:border-none"
          />
        </div>
      </div>

      {/* Contact List */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto im-scroll relative">
        {/* Quick Access Groups — TG light-blue icons */}
        <div className="pb-1">
          {contactGroups.map((group) => (
            <div
              key={group.title}
              className="im-contact-item"
              onClick={() => {
                if (group.title === '新的朋友') {
                  setShowFriendRequests(true);
                }
                if (group.title === '群聊') {
                  setShowGroupPanel(true);
                }
              }}
            >
              <div
                className="flex items-center justify-center shrink-0"
                style={{
                  width: '44px',
                  height: '44px',
                  borderRadius: '12px',
                  backgroundColor: group.title === '新的朋友'
                    ? 'rgba(51,144,236,0.1)'
                    : 'rgba(51,144,236,0.1)',
                }}
              >
                <span style={{ color: '#3390EC' }}>{iconMap[group.icon]}</span>
              </div>
              <div className="flex-1 min-w-0">
                <span className="text-sm" style={{ color: '#1C2733' }}>{group.title}</span>
              </div>
              {group.title === '新的朋友' && friendRequestUnreadCount > 0 && (
                <span
                  className="flex items-center justify-center shrink-0"
                  style={{
                    minWidth: '20px',
                    height: '20px',
                    borderRadius: '10px',
                    backgroundColor: '#E53935',
                    color: '#FFFFFF',
                    fontSize: '11px',
                    fontWeight: 700,
                    padding: '0 5px',
                  }}
                >
                  {friendRequestUnreadCount > 99 ? '99+' : friendRequestUnreadCount}
                </span>
              )}
              {group.title === '群聊' && groupAppUnreadCount > 0 && (
                <span
                  className="flex items-center justify-center shrink-0"
                  style={{
                    minWidth: '20px',
                    height: '20px',
                    borderRadius: '10px',
                    backgroundColor: '#E53935',
                    color: '#FFFFFF',
                    fontSize: '11px',
                    fontWeight: 700,
                    padding: '0 5px',
                  }}
                >
                  {groupAppUnreadCount > 99 ? '99+' : groupAppUnreadCount}
                </span>
              )}
              <ChevronRight
                className="w-4 h-4 shrink-0"
                style={{ color: '#C4C9CC' }}
              />
            </div>
          ))}
        </div>

        {/* Alphabetical Contact Groups */}
        {!searchQuery && groupedContacts.map(([letter, list]) => (
          <div key={letter} id={`letter-${letter}`}>
            {/* Sticky alphabet header — TG subtle blue */}
            <div
              className="sticky top-0 z-[1]"
              style={{
                padding: '6px 16px',
                fontSize: '13px',
                fontWeight: 600,
                color: '#3390EC',
                backgroundColor: 'rgba(51,144,236,0.06)',
              }}
            >
              {letter}
            </div>
            {list.map((contact) => (
              <div key={contact.id} className={`im-contact-item${selectedContactId === contact.id ? ' selected' : ''}`} onClick={() => handleContactClick(contact)}>
                <div className="relative shrink-0">
                  <Avatar
                    style={{ width: '44px', height: '44px', borderRadius: '50%' }}
                  >
                    <AvatarImage src={contact.avatar} alt={contact.name} />
                    <AvatarFallback
                      className="text-sm"
                      style={{
                        backgroundColor: '#E8EDEF',
                        color: '#708499',
                        fontSize: '15px',
                        fontWeight: 500,
                      }}
                    >
                      {contact.name[0]}
                    </AvatarFallback>
                  </Avatar>
                  {contact.online && (
                    <span
                      className="absolute bottom-0 right-0 rounded-full"
                      style={{
                        width: '12px',
                        height: '12px',
                        backgroundColor: '#4DCD5E',
                        border: '2.5px solid #FFFFFF',
                      }}
                    />
                  )}
                </div>
                <span className="text-sm" style={{ color: '#1C2733' }}>{contact.name}</span>
              </div>
            ))}
          </div>
        ))}

        {/* Search Results */}
        {searchQuery && (
          <div>
            {filteredContacts.length > 0 ? (
              filteredContacts.map((contact) => (
                <div key={contact.id} className={`im-contact-item${selectedContactId === contact.id ? ' selected' : ''}`} onClick={() => handleContactClick(contact)}>
                  <div className="relative shrink-0">
                    <Avatar
                      style={{ width: '44px', height: '44px', borderRadius: '50%' }}
                    >
                      <AvatarImage src={contact.avatar} alt={contact.name} />
                      <AvatarFallback
                        className="text-sm"
                        style={{
                          backgroundColor: '#E8EDEF',
                          color: '#708499',
                          fontSize: '15px',
                          fontWeight: 500,
                        }}
                      >
                        {contact.name[0]}
                      </AvatarFallback>
                    </Avatar>
                    {contact.online && (
                      <span
                        className="absolute bottom-0 right-0 rounded-full"
                        style={{
                          width: '12px',
                          height: '12px',
                          backgroundColor: '#4DCD5E',
                          border: '2.5px solid #FFFFFF',
                        }}
                      />
                    )}
                  </div>
                  <span className="text-sm" style={{ color: '#1C2733' }}>{contact.name}</span>
                </div>
              ))
            ) : (
              <div
                className="flex items-center justify-center h-32 text-sm"
                style={{ color: '#A2ACB5' }}
              >
                没有找到联系人
              </div>
            )}
          </div>
        )}

        {/* Total count — clean muted text */}
        {!searchQuery && (
          <div
            className="text-center"
            style={{
              padding: '12px 16px',
              fontSize: '13px',
              color: '#A2ACB5',
            }}
          >
            {totalContacts} 位联系人
          </div>
        )}

        {/* Alphabet Index (right side) */}
        {!searchQuery && (
          <div className="im-alpha-index">
            {alphabetIndex.map(letter => (
              <span key={letter} onClick={() => scrollToLetter(letter)}>{letter}</span>
            ))}
          </div>
        )}
      </div>


    </div>
  );
}
